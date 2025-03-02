package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"practicum_final_project/database"
	"practicum_final_project/handlers"
)

func main() {
	// Получаем рабочий каталог
	appPath, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	// Формируем путь к базе данных (scheduler.db должен находиться в корне проекта)
	dbPath := filepath.Join(appPath, "scheduler.db")

	// Инициализируем БД
	if err := database.Init(dbPath); err != nil {
		log.Fatalf("Ошибка инициализации базы данных: %v", err)
	}
	defer database.DB.Close()

	// Инициализируем handlers с подключением к БД
	handlers.InitDB(database.DB)

	// Формируем путь к директории web (корень проекта)
	webDir := filepath.Join(appPath, "web")
	if _, err := os.Stat(webDir); err != nil {
		log.Fatalf("Ошибка: директория web не найдена")
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
