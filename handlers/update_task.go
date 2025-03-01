package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"practicum_final_project/models"
	"practicum_final_project/utils"
)

// HandleUpdateTask обновляет существующую задачу.
func HandleUpdateTask(w http.ResponseWriter, r *http.Request) {
	var task models.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		utils.RespondError(w, "Неверный формат JSON", http.StatusBadRequest)
		return
	}

	if task.ID == "" {
		utils.RespondError(w, "Не указан идентификатор", http.StatusBadRequest)
		return
	}
	if task.Title == "" {
		utils.RespondError(w, "Заголовок задачи обязателен", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(task.ID, 10, 64)
	if err != nil {
		utils.RespondError(w, "Некорректный идентификатор", http.StatusBadRequest)
		return
	}

	var exists bool
	err = DB.QueryRow("SELECT EXISTS(SELECT 1 FROM scheduler WHERE id = ?)", id).Scan(&exists)
	if err != nil {
		utils.RespondError(w, "Ошибка базы данных", http.StatusInternalServerError)
		return
	}
	if !exists {
		utils.RespondError(w, "Задача не найдена", http.StatusNotFound)
		return
	}

	now := time.Now().UTC().Truncate(24 * time.Hour)
	parsedDate, err := time.Parse("20060102", task.Date)
	if err != nil {
		utils.RespondError(w, "Неверный формат даты", http.StatusBadRequest)
		return
	}

	if task.Repeat != "" {
		nextDate, err := utils.NextDate(now, task.Date, task.Repeat)
		if err != nil {
			utils.RespondError(w, err.Error(), http.StatusBadRequest)
			return
		}
		task.Date = nextDate
	} else if parsedDate.Before(now) {
		// Если правило повторения не указано, выдаём ошибку.
		utils.RespondError(w, "Дата не может быть меньше сегодняшней", http.StatusBadRequest)
		return
	}

	_, err = DB.Exec(
		`UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?`,
		task.Date, task.Title, task.Comment, task.Repeat, id,
	)
	if err != nil {
		utils.RespondError(w, "Ошибка базы данных: "+err.Error(), http.StatusInternalServerError)
		return
	}

	utils.RespondJSON(w, map[string]interface{}{})
}
