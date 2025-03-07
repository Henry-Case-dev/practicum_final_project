package handlers

import (
	"database/sql"
)

// DB – глобальная переменная для доступа к базе данных из обработчиков.
var DB *sql.DB

// InitDB инициализирует доступ к базе данных для хендлеров.
func InitDB(database *sql.DB) {
	DB = database
}
