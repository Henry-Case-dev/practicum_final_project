package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Определяем порт
	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = "7540"
	}

	// Определяем путь к БД
	dbPath := os.Getenv("TODO_DBFILE")
	if dbPath == "" {
		dbPath = "scheduler.db"
	}

	// Проверяем существование файла БД
	_, err := os.Stat(dbPath)
	installDB := os.IsNotExist(err)

	// Открываем соединение с БД
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Создаём таблицу при первом запуске
	if installDB {
		if err := createTables(db); err != nil {
			log.Fatal("Ошибка создания БД:", err)
		}
		log.Println("База данных инициализирована")
	}

	// Настраиваем файловый сервер
	webDir := "./web"
	http.Handle("/", http.FileServer(http.Dir(webDir)))

	// Запускаем сервер
	log.Printf("Сервер запущен на порту %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func createTables(db *sql.DB) error {
	// Создаём таблицу задач
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
