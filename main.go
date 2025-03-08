package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"practicum_final_project/database"
	"practicum_final_project/handlers"
	"practicum_final_project/utils"
)

func main() {
	// НЕ устанавливаем глобальный часовой пояс - используем AppTimeZone в utils
	// time.Local = time.UTC

	// Логируем информацию о системном времени и часовом поясе
	log.Printf("Системное время: %v, Часовой пояс: %v, UTC: %v, AppTimeZone: %v",
		time.Now(), time.Now().Location(), time.Now().UTC(), time.Now().In(utils.AppTimeZone))

	// Получаем путь к исполняемому файлу
	exePath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}

	// Формируем путь к базе данных (scheduler.db должен находиться в родительской директории исполняемого файла)
	dbPath := filepath.Join(filepath.Dir(exePath), "scheduler.db")

	// Инициализируем БД
	if err := database.Init(dbPath); err != nil {
		log.Fatalf("Ошибка инициализации базы данных: %v", err)
	}
	defer database.DB.Close()

	// Инициализируем handlers с подключением к БД
	handlers.InitDB(database.DB)

	// Формируем путь к директории web (находится в родительской директории исполняемого файла)
	webDir := filepath.Join(filepath.Dir(exePath), "web")
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

	// Определяем порт
	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = "7540"
	}
	log.Printf("Сервер запущен на порту %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
