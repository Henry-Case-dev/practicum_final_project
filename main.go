package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"practicum_final_project/handlers"
	"practicum_final_project/utils"

	_ "github.com/mattn/go-sqlite3"
)

// db - глобальная переменная для подключения к базе данных
var db *sql.DB

func main() {
	// Получаем путь к базе данных
	dbPath := utils.GetDBPath()
	var err error
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Инициализируем базу данных в пакете handlers
	handlers.InitDB(db)

	// Создаем таблицы, если они не существуют
	if err := utils.CreateTables(db); err != nil {
		log.Fatal("Failed to create tables:", err)
	}

	// Получаем порт для сервера из переменной окружения
	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = "7540"
	}

	// Настраиваем маршруты для HTTP-запросов
	http.Handle("/", http.FileServer(http.Dir("./web")))
	http.HandleFunc("/api/nextdate", handlers.HandleNextDate)
	http.HandleFunc("/api/task", handlers.HandleTask)
	http.HandleFunc("/api/tasks", handlers.HandleTasks)
	http.HandleFunc("/api/task/done", handlers.HandleTaskDone)
	http.HandleFunc("/api/task/delete", handlers.HandleDeleteTask)

	// Запускаем сервер
	log.Printf("Server started on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
