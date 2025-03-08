package utils

import (
	"database/sql"
	"os"
	"path/filepath"
)

func GetDBPath() string {
	if customPath := os.Getenv("TODO_DBFILE"); customPath != "" {
		return customPath
	}
	return filepath.Join(".", "scheduler.db")
}

func CreateTables(db *sql.DB) error {
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
