package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Task struct {
	ID      int64  `db:"id"`
	Date    string `db:"date"`
	Title   string `db:"title"`
	Comment string `db:"comment"`
	Repeat  string `db:"repeat"`
}

func createTables(db *sql.DB) error {
	query := `
	CREATE TABLE scheduler (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		date TEXT NOT NULL,
		title TEXT NOT NULL,
		comment TEXT,
		repeat TEXT
	);
	CREATE INDEX idx_date ON scheduler (date);`

	_, err := db.Exec(query)
	return err
}

func NextDate(now time.Time, dateStr, repeat string) (string, error) {
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
				break
			}
		}
		return nextDate.Format("20060102"), nil

	case repeat == "y":
		nextDate := parsedDate.AddDate(1, 0, 0)
		if parsedDate.Month() == time.February && parsedDate.Day() == 29 {
			if !isLeap(nextDate.Year()) {
				nextDate = time.Date(nextDate.Year(), time.March, 1, 0, 0, 0, 0, nextDate.Location())
			}
		}
		return nextDate.Format("20060102"), nil

	default:
		if repeat != "" {
			return "", fmt.Errorf("unsupported repeat rule")
		}
		return "", nil
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

	nowStr := r.FormValue("now")
	dateStr := r.FormValue("date")
	repeat := r.FormValue("repeat")

	if nowStr == "" || dateStr == "" || repeat == "" {
		respondError(w, "missing parameters", http.StatusBadRequest)
		return
	}

	now, err := time.Parse("20060102", nowStr)
	if err != nil {
		respondError(w, "invalid 'now' parameter", http.StatusBadRequest)
		return
	}

	nextDate, err := NextDate(now, dateStr, repeat)
	if err != nil {
		respondError(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(nextDate))
}

func handleTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var taskInput struct {
		Date    string `json:"date"`
		Title   string `json:"title"`
		Comment string `json:"comment"`
		Repeat  string `json:"repeat"`
	}

	if err := json.NewDecoder(r.Body).Decode(&taskInput); err != nil {
		respondError(w, "Ошибка десериализации JSON", http.StatusBadRequest)
		return
	}

	if taskInput.Title == "" {
		respondError(w, "Не указан заголовок задачи", http.StatusBadRequest)
		return
	}

	// Основная логика обработки даты
	now := time.Now().UTC()
	today := now.Format("20060102")
	finalDate := taskInput.Date

	// Если дата не указана - использовать сегодня
	if finalDate == "" {
		finalDate = today
	}

	// Парсим и проверяем формат даты
	parsedDate, err := time.Parse("20060102", finalDate)
	if err != nil {
		respondError(w, "invalid date format", http.StatusBadRequest)
		return
	}

	// Проверяем правило повторения
	if taskInput.Repeat != "" {
		_, err = NextDate(now, finalDate, taskInput.Repeat)
		if err != nil {
			respondError(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	// Корректируем дату если она в прошлом
	if parsedDate.Before(now) {
		if taskInput.Repeat == "" {
			finalDate = today
		} else {
			next, err := NextDate(now, finalDate, taskInput.Repeat)
			if err != nil {
				respondError(w, err.Error(), http.StatusBadRequest)
				return
			}
			finalDate = next
		}
	}

	// Подключаемся к БД
	db, err := sql.Open("sqlite3", getDBPath())
	if err != nil {
		respondError(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Вставляем задачу
	res, err := db.Exec(
		`INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`,
		finalDate,
		taskInput.Title,
		taskInput.Comment,
		taskInput.Repeat,
	)
	if err != nil {
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, _ := res.LastInsertId()
	respondJSON(w, map[string]string{"id": strconv.FormatInt(id, 10)})
}

func respondJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func getDBPath() string {
	if path := os.Getenv("TODO_DBFILE"); path != "" {
		return path
	}
	return "scheduler.db"
}

func main() {
	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = "7540"
	}

	// Инициализация БД
	db, err := sql.Open("sqlite3", getDBPath())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if _, err := os.Stat(getDBPath()); os.IsNotExist(err) {
		if err := createTables(db); err != nil {
			log.Fatal("Ошибка создания БД:", err)
		}
		log.Println("База данных инициализирована")
	}

	// Роутинг
	http.Handle("/", http.FileServer(http.Dir("./web")))
	http.HandleFunc("/api/nextdate", handleNextDate)
	http.HandleFunc("/api/task", handleTask)

	log.Printf("Сервер запущен на порту %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
