package database

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func Init(dbPath string) error {
	var err error
	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}

	// Проверяем соединение с базой данных
	if err = DB.Ping(); err != nil {
		return err
	}

	// Создаем таблицу если её нет
	schema := `
    CREATE TABLE IF NOT EXISTS scheduler (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        date TEXT NOT NULL,
        title TEXT NOT NULL,
        comment TEXT,
        repeat TEXT
    );
    CREATE INDEX IF NOT EXISTS idx_date ON scheduler (date);`

	_, err = DB.Exec(schema)
	if err != nil {
		log.Printf("Ошибка создания таблицы: %v", err)
		return err
	}

	return nil
}
