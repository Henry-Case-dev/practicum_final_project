package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"practicum_final_project/database"
	"practicum_final_project/handlers"
	"practicum_final_project/tests"
)

func main() {
	// Используем путь к БД из settings.go
	dbPath := tests.DBFile
	if envPath := os.Getenv("TODO_DBFILE"); envPath != "" {
		dbPath = envPath
	}

	// Инициализируем БД
	if err := database.Init(dbPath); err != nil {
		log.Fatalf("Ошибка инициализации базы данных: %v", err)
	}
	defer database.DB.Close()

	// Инициализируем handlers
	handlers.InitDB(database.DB)

	// Определяем директорию web относительно текущей
	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	// Путь к директории web
	webDir := filepath.Join(currentDir, "web")
	if _, err := os.Stat(webDir); err != nil {
		// Если web не найдена в текущей директории, ищем в родительской
		webDir = filepath.Join(filepath.Dir(currentDir), "web")
		if _, err := os.Stat(webDir); err != nil {
			log.Fatalf("Ошибка: директория web не найдена")
		}
	}

	// Файловый сервер для фронтенда
	http.Handle("/", http.FileServer(http.Dir(webDir)))

	// Регистрируем API-обработчики
	http.HandleFunc("/api/task", handlers.HandleTask)
	http.HandleFunc("/api/tasks", handlers.HandleTasks)
	http.HandleFunc("/api/nextdate", handlers.HandleNextDate)
	http.HandleFunc("/api/task/done", handlers.HandleTaskDone)
	http.HandleFunc("/api/task/delete", handlers.HandleDeleteTask)

	// Определяем порт
	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = "7540"
	}
	log.Printf("Сервер запущен на порту %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
