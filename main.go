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
	ID      int64  `json:"id"`
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
		for {
			nextDate = nextDate.AddDate(0, 0, days)
			if nextDate.After(now) {
				return nextDate.Format("20060102"), nil
			}
		}

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

func handleTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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
	if task.Date == "" || task.Date == "today" {
		task.Date = now.Format("20060102")
	}

	parsedDate, err := time.Parse("20060102", task.Date)
	if err != nil {
		respondError(w, "Invalid date format", http.StatusBadRequest)
		return
	}

	//
	if parsedDate.Before(now) {
		task.Date = now.Format("20060102")
		parsedDate = now
	}

	//
	if !parsedDate.Before(now) && task.Repeat != "" {
		next, err := NextDate(now, task.Date, task.Repeat)
		if err != nil {
			respondError(w, err.Error(), http.StatusBadRequest)
			return
		}
		if next != "" {
			task.Date = next
		}
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
	respondJSON(w, map[string]int64{"id": id})
}

func handleNextDate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	nowStr := r.URL.Query().Get("now")
	dateStr := r.URL.Query().Get("date")
	repeat := r.URL.Query().Get("repeat")

	now, err := time.Parse("20060102", nowStr)
	if err != nil {
		now = time.Now().UTC()
	}

	nextDate, err := NextDate(now, dateStr, repeat)
	if err != nil {
		respondError(w, err.Error(), http.StatusBadRequest)
		return
	}

	respondJSON(w, map[string]string{"date": nextDate})
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

	appPath, _ := os.Getwd()
	return filepath.Join(appPath, "scheduler.db")
}

func main() {
	// Инициализация БД
	var err error
	dbPath := getDBPath()
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := createTables(); err != nil {
		log.Fatal("Failed to create tables:", err)
	}

	// Настройка сервера
	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = "7540"
	}

	http.Handle("/", http.FileServer(http.Dir("./web")))
	http.HandleFunc("/api/nextdate", handleNextDate)
	http.HandleFunc("/api/task", handleTask)

	log.Printf("Server started on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
