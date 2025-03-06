package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"practicum_final_project/models"
	"practicum_final_project/utils"
)

// HandleCreateTask обрабатывает создание новой задачи.
// Если поле date отсутствует или пустое, то используется текущее время (формат "20060102").
// При наличии правила повторения поле date должно быть равно сегодняшней, иначе возвращается ошибка.
// Для не повторяющихся задач, если переданная дата меньше сегодняшней, возвращается ошибка.
func HandleCreateTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.RespondError(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var task models.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		utils.RespondError(w, "Неверный формат JSON", http.StatusBadRequest)
		return
	}
	if task.Title == "" {
		utils.RespondError(w, "Не указан заголовок задачи", http.StatusBadRequest)
		return
	}

	// Используем текущее локальное время, обрезанное до начала дня, как today.
	today := time.Now().Local().Truncate(24 * time.Hour)
	log.Printf("DEBUG (CreateTask): today=%s", today.Format("20060102"))

	// Обработка даты:
	if task.Date == "" {
		task.Date = today.Format("20060102")
		log.Printf("DEBUG (CreateTask): date not provided, set to today=%s", task.Date)
	} else {
		parsedDate, err := time.Parse("20060102", task.Date)
		if err != nil {
			utils.RespondError(w, "Неверный формат даты", http.StatusBadRequest)
			return
		}
		if task.Repeat != "" {
			// Для повторяющихся задач date должно совпадать с today.
			if task.Date != today.Format("20060102") {
				log.Printf("DEBUG (CreateTask): repeat task but date (%s) != today (%s)", task.Date, today.Format("20060102"))
				utils.RespondError(w, "Дата должна быть сегодняшняя", http.StatusBadRequest)
				return
			}
		} else {
			// Для не повторяющихся задач, если дата меньше today, возвращаем ошибку.
			if parsedDate.Before(today) {
				log.Printf("DEBUG (CreateTask): non-repeat task but date (%s) < today (%s)", task.Date, today.Format("20060102"))
				utils.RespondError(w, "Дата не может быть меньше сегодняшней", http.StatusBadRequest)
				return
			}
		}
	}

	result, err := DB.Exec(
		`INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`,
		task.Date, task.Title, task.Comment, task.Repeat,
	)
	if err != nil {
		utils.RespondError(w, "Ошибка базы данных: "+err.Error(), http.StatusInternalServerError)
		return
	}
	id, err := result.LastInsertId()
	if err != nil {
		utils.RespondError(w, "Не удалось получить id задачи", http.StatusInternalServerError)
		return
	}
	log.Printf("DEBUG (CreateTask): task created with id=%d, date=%s", id, task.Date)

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	utils.RespondJSON(w, map[string]string{"id": strconv.FormatInt(id, 10)})
}
