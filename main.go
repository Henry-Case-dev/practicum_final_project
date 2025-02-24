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

// Task представляет структуру задачи
type Task struct {
	ID      int64  `json:"id"`      // Уникальный идентификатор задачи
	Date    string `json:"date"`    // Дата выполнения задачи в формате YYYYMMDD
	Title   string `json:"title"`   // Заголовок задачи
	Comment string `json:"comment"` // Комментарий к задаче
	Repeat  string `json:"repeat"`  // Правило повторения задачи (d N - каждый N дней, y - каждый год)
}

// Глобальная переменная для хранения экземпляра БД
var db *sql.DB

// createTables создает таблицы в базе данных, если они не существуют
func createTables() error {
	query := `CREATE TABLE IF NOT EXISTS scheduler (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        date TEXT NOT NULL,
        title TEXT NOT NULL,
        comment TEXT,
        repeat TEXT
    );
    CREATE INDEX IF NOT EXISTS idx_date ON scheduler (date);` // Создаем индекс для быстрого поиска задач по дате

	_, err := db.Exec(query) // Выполняем SQL-запрос
	return err
}

// NextDate вычисляет следующую дату выполнения задачи на основе правила повторения
func NextDate(now time.Time, dateStr, repeat string) (string, error) {
	// Если правило повторения не задано, возвращаем пустую строку и nil error
	if repeat == "" {
		return "", nil
	}

	// Парсим строку с датой в объект time.Time
	parsedDate, err := time.Parse("20060102", dateStr)
	if err != nil {
		return "", fmt.Errorf("invalid date: %v", err) // Возвращаем ошибку, если дата имеет неверный формат
	}

	// Обрабатываем правило повторения
	switch {
	// Если правило начинается с "d " (каждые несколько дней)
	case strings.HasPrefix(repeat, "d "):
		parts := strings.Split(repeat, " ") // Разбиваем строку на части по пробелу
		if len(parts) != 2 {
			return "", fmt.Errorf("invalid repeat format") // Возвращаем ошибку, если формат правила неверен
		}

		days, err := strconv.Atoi(parts[1]) // Преобразуем количество дней в число
		if err != nil || days < 1 || days > 400 {
			return "", fmt.Errorf("invalid days value") // Возвращаем ошибку, если количество дней не является числом или находится вне допустимого диапазона
		}

		// Вычисляем следующую дату, добавляя дни до тех пор, пока она не станет больше или равна текущей дате
		nextDate := parsedDate
		for nextDate.Before(now) || nextDate.Equal(now) {
			nextDate = nextDate.AddDate(0, 0, days) // Добавляем указанное количество дней
		}
		return nextDate.Format("20060102"), nil // Возвращаем следующую дату в формате YYYYMMDD

	// Если правило равно "y" (каждый год)
	case repeat == "y":
		nextDate := parsedDate
		// Вычисляем следующую дату, добавляя год до тех пор, пока она не станет больше текущей даты
		for {
			nextDate = nextDate.AddDate(1, 0, 0) // Добавляем один год

			// Корректируем дату, если исходная дата - 29 февраля невисокосного года
			if parsedDate.Month() == time.February && parsedDate.Day() == 29 {
				if !isLeap(nextDate.Year()) {
					nextDate = nextDate.AddDate(0, 0, 1) // Переносим на 1 марта
				}
			}
			if nextDate.After(now) {
				return nextDate.Format("20060102"), nil // Возвращаем следующую дату в формате YYYYMMDD
			}
		}

	// Если правило не поддерживается
	default:
		return "", fmt.Errorf("unsupported repeat rule") // Возвращаем ошибку
	}
}

// isLeap проверяет, является ли год високосным
func isLeap(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

// handleTask обрабатывает запросы к эндпоинту /api/task (добавление задач)
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

	// Всегда получаем актуальное время для каждой задачи
	now := time.Now().UTC()
	today := now.Format("20060102")

	// Обработка даты задачи
	if task.Date == "" || task.Date == "today" {
		task.Date = today
	}

	parsedDate, err := time.Parse("20060102", task.Date)
	if err != nil {
		respondError(w, "Invalid date format", http.StatusBadRequest)
		return
	}

	// Нормализация дат до начала суток
	nowDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	taskDate := parsedDate.UTC().Truncate(24 * time.Hour)

	// Корректировка даты только если задача без повтора и дата в прошлом
	if task.Repeat == "" && taskDate.Before(nowDate) {
		task.Date = today
	} else if task.Repeat != "" {
		// Для повторяющихся задач всегда вычисляем следующую дату
		next, err := NextDate(now, task.Date, task.Repeat)
		if err != nil {
			respondError(w, err.Error(), http.StatusBadRequest)
			return
		}
		task.Date = next
	}

	// Вставка в БД
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

// handleNextDate обрабатывает запросы к эндпоинту /api/nextdate (вычисление следующей даты)
func handleNextDate(w http.ResponseWriter, r *http.Request) {
	// Проверяем метод запроса (должен быть GET)
	if r.Method != http.MethodGet {
		respondError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Получаем параметры запроса из URL
	nowStr := r.URL.Query().Get("now")    // Текущая дата
	dateStr := r.URL.Query().Get("date")  // Дата задачи
	repeat := r.URL.Query().Get("repeat") // Правило повторения

	// Если не удалось распарсить дату, устанавливаем текущую дату
	now, err := time.Parse("20060102", nowStr)
	if err != nil {
		now = time.Now().UTC()
	}

	// Вычисляем следующую дату с помощью функции NextDate
	nextDate, err := NextDate(now, dateStr, repeat)
	if err != nil {
		respondError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Отправляем JSON-ответ со следующей датой
	respondJSON(w, map[string]string{"date": nextDate})
}

// respondJSON отправляет JSON-ответ
func respondJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// respondError отправляет JSON-ответ с ошибкой
func respondError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// getDBPath определяет путь к файлу базы данных
func getDBPath() string {
	// Если переменная окружения TODO_DBFILE установлена, используем её
	if customPath := os.Getenv("TODO_DBFILE"); customPath != "" {
		return customPath
	}

	// Иначе используем путь по умолчанию (scheduler.db в текущей директории)
	appPath, _ := os.Getwd()
	return filepath.Join(appPath, "scheduler.db")
}

// main - основная функция приложения
func main() {
	// Инициализация БД
	var err error
	dbPath := getDBPath()                 // Получаем путь к файлу БД
	db, err = sql.Open("sqlite3", dbPath) // Открываем соединение с БД SQLite3
	if err != nil {
		log.Fatal(err) // Если не удалось открыть соединение, завершаем программу с ошибкой
	}
	defer db.Close() // Закрываем соединение с БД при завершении работы программы

	// Создаем таблицы, если они не существуют
	if err := createTables(); err != nil {
		log.Fatal("Failed to create tables:", err) // Если не удалось создать таблицы, завершаем программу с ошибкой
	}

	// Настройка сервера
	port := os.Getenv("TODO_PORT") // Получаем номер порта из переменной окружения TODO_PORT
	if port == "" {
		port = "7540" // Если переменная не установлена, используем порт 7540 по умолчанию
	}

	// Регистрируем обработчики HTTP
	http.Handle("/", http.FileServer(http.Dir("./web"))) // Файловый сервер для статических файлов (HTML, CSS, JS)
	http.HandleFunc("/api/nextdate", handleNextDate)     // Обработчик для эндпоинта /api/nextdate
	http.HandleFunc("/api/task", handleTask)             // Обработчик для эндпоинта /api/task

	// Запускаем HTTP-сервер
	log.Printf("Server started on port %s", port) // Выводим сообщение о запуске сервера
	log.Fatal(http.ListenAndServe(":"+port, nil)) // Запускаем сервер и ждем входящие соединения. Если произошла ошибка, завершаем программу.
}
