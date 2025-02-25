package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

var db *sql.DB

func createTables() error {
	query := `CREATE TABLE IF NOT EXISTS scheduler (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		date TEXT NOT NULL,
		title TEXT NOT NULL,
		comment TEXT,
		repeat TEXT
	);
	CREATE INDEX IF NOT EXISTS idx_date ON scheduler (date);`

	_, err := db.Exec(query)
	return err
}

func NextDate(now time.Time, dateStr, repeat string) (string, error) {
	if repeat == "" {
		return "", nil
	}

	parsedDate, err := time.Parse("20060102", dateStr)
	if err != nil {
		return "", fmt.Errorf("invalid date: %v", err)
	}

	switch {
	case strings.HasPrefix(repeat, "d "):
		parts := strings.Split(repeat, " ")
		if len(parts) != 2 {
			return "", fmt.Errorf("invalid repeat format")
		}

		days, err := strconv.Atoi(parts[1])
		if err != nil || days < 1 || days > 400 {
			return "", fmt.Errorf("invalid days value")
		}

		nextDate := parsedDate
		for nextDate.Before(now) || nextDate.Equal(now) {
			nextDate = nextDate.AddDate(0, 0, days)
		}
		return nextDate.Format("20060102"), nil

	case repeat == "y":
		nextDate := parsedDate
		for {
			nextDate = nextDate.AddDate(1, 0, 0)
			if parsedDate.Month() == time.February && parsedDate.Day() == 29 {
				if !isLeap(nextDate.Year()) {
					nextDate = nextDate.AddDate(0, 0, 1)
				}
			}
			if nextDate.After(now) {
				return nextDate.Format("20060102"), nil
			}
		}

	default:
		return "", fmt.Errorf("unsupported repeat rule")
	}
}
func isLeap(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

func handleNextDate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	date := r.URL.Query().Get("date")
	repeat := r.URL.Query().Get("repeat")

	if date == "" || repeat == "" {
		respondError(w, "Missing date or repeat parameter", http.StatusBadRequest)
		return
	}

	now := time.Now().UTC()
	nextDate, err := NextDate(now, date, repeat)
	if err != nil {
		respondError(w, err.Error(), http.StatusBadRequest)
		return
	}

	respondJSON(w, map[string]string{"date": nextDate})
}

func handleTask(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		handleCreateTask(w, r)
	case http.MethodGet, http.MethodPut:
		handleTaskRoutes(w, r)
	default:
		respondError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleCreateTask(w http.ResponseWriter, r *http.Request) {
	var task Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		respondError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if task.Title == "" {
		respondError(w, "Title is required", http.StatusBadRequest)
		return
	}

	now := time.Now().UTC()
	if task.Date == "" {
		task.Date = now.Format("20060102")
	}

	parsedDate, err := time.Parse("20060102", task.Date)
	if err != nil {
		respondError(w, "Invalid date format", http.StatusBadRequest)
		return
	}

	if task.Repeat != "" {
		next, err := NextDate(now, task.Date, task.Repeat)
		if err != nil {
			respondError(w, err.Error(), http.StatusBadRequest)
			return
		}
		task.Date = next
	} else if parsedDate.Before(now) {
		task.Date = now.Format("20060102")
	}

	res, err := db.Exec(
		`INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`,
		task.Date, task.Title, task.Comment, task.Repeat,
	)
	if err != nil {
		respondError(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	id, _ := res.LastInsertId()
	respondJSON(w, map[string]string{"id": strconv.FormatInt(id, 10)})
}
func handleTaskRoutes(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handleGetTask(w, r)
	case http.MethodPut:
		handleUpdateTask(w, r)
	default:
		respondError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleGetTask(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		respondError(w, "Не указан идентификатор", http.StatusBadRequest)
		return
	}

	var task Task
	row := db.QueryRow(
		`SELECT id, date, title, comment, repeat 
		FROM scheduler WHERE id = ?`,
		id,
	)

	var dbID int64
	err := row.Scan(&dbID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		if err == sql.ErrNoRows {
			respondError(w, "Задача не найдена", http.StatusNotFound)
		} else {
			respondError(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	task.ID = strconv.FormatInt(dbID, 10)
	respondJSON(w, task)
}

func handleUpdateTask(w http.ResponseWriter, r *http.Request) {
	var task Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		respondError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if task.ID == "" {
		respondError(w, "Не указан идентификатор", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(task.ID, 10, 64)
	if err != nil {
		respondError(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	var exists bool
	db.QueryRow("SELECT EXISTS(SELECT 1 FROM scheduler WHERE id = ?)", id).Scan(&exists)
	if !exists {
		respondError(w, "Задача не найдена", http.StatusNotFound)
		return
	}

	now := time.Now().UTC()
	if task.Title == "" {
		respondError(w, "Title is required", http.StatusBadRequest)
		return
	}

	parsedDate, err := time.Parse("20060102", task.Date)
	if err != nil {
		respondError(w, "Invalid date format", http.StatusBadRequest)
		return
	}

	if task.Repeat != "" {
		next, err := NextDate(now, task.Date, task.Repeat)
		if err != nil {
			respondError(w, err.Error(), http.StatusBadRequest)
			return
		}
		task.Date = next
	} else if parsedDate.Before(now) {
		task.Date = now.Format("20060102")
	}

	_, err = db.Exec(
		`UPDATE scheduler 
		SET date = ?, title = ?, comment = ?, repeat = ?
		WHERE id = ?`,
		task.Date, task.Title, task.Comment, task.Repeat, id,
	)

	if err != nil {
		respondError(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, map[string]interface{}{})
}

func handleTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	search := r.URL.Query().Get("search")
	limit := 50

	var rows *sql.Rows
	var err error

	if search != "" {
		if date, err := time.Parse("02.01.2006", search); err == nil {
			rows, err = db.Query(
				`SELECT id, date, title, comment, repeat 
				FROM scheduler 
				WHERE date = ? 
				ORDER BY date 
				LIMIT ?`,
				date.Format("20060102"),
				limit,
			)
		} else {
			searchPattern := "%" + search + "%"
			rows, err = db.Query(
				`SELECT id, date, title, comment, repeat 
				FROM scheduler 
				WHERE title LIKE ? OR comment LIKE ? 
				ORDER BY date 
				LIMIT ?`,
				searchPattern,
				searchPattern,
				limit,
			)
		}
	} else {
		rows, err = db.Query(
			`SELECT id, date, title, comment, repeat 
			FROM scheduler 
			ORDER BY date 
			LIMIT ?`,
			limit,
		)
	}

	if err != nil {
		respondError(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	tasks := make([]Task, 0)
	for rows.Next() {
		var id int64
		var t Task
		err := rows.Scan(&id, &t.Date, &t.Title, &t.Comment, &t.Repeat)
		if err != nil {
			respondError(w, "Error reading tasks", http.StatusInternalServerError)
			return
		}
		t.ID = strconv.FormatInt(id, 10)
		tasks = append(tasks, t)
	}

	if tasks == nil {
		tasks = make([]Task, 0)
	}

	respondJSON(w, map[string][]Task{"tasks": tasks})
}

func respondJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func getDBPath() string {
	if customPath := os.Getenv("TODO_DBFILE"); customPath != "" {
		return customPath
	}
	return filepath.Join(".", "scheduler.db")
}

func main() {
	dbPath := getDBPath()
	var err error
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := createTables(); err != nil {
		log.Fatal("Failed to create tables:", err)
	}

	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = "7540"
	}

	http.Handle("/", http.FileServer(http.Dir("./web")))
	http.HandleFunc("/api/nextdate", handleNextDate)
	http.HandleFunc("/api/task", handleTask)
	http.HandleFunc("/api/tasks", handleTasks)

	log.Printf("Server started on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
