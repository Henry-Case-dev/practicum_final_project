package handlers

import (
	"database/sql"
)

var DB *sql.DB

// InitDB инициализирует соединение с БД для хендлеров.
func InitDB(database *sql.DB) {
	DB = database
}
