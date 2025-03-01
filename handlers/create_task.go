package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"practicum_final_project/models"
	"practicum_final_project/utils"
)

// HandleCreateTask обрабатывает создание новой задачи.
func HandleCreateTask(w http.ResponseWriter, r *http.Request) {
	var task models.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		utils.RespondError(w, "Неверный формат JSON", http.StatusBadRequest)
		return
	}

	if task.Title == "" {
		utils.RespondError(w, "Заголовок задачи обязателен", http.StatusBadRequest)
		return
	}

	now := time.Now().UTC().Truncate(24 * time.Hour)
	// Если дата не указана, используем сегодняшнюю.
	if task.Date == "" {
		task.Date = now.Format("20060102")
	}

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
		// Если правило повторения не указано, выдаём ошибку, если дата меньше сегодняшней.
		utils.RespondError(w, "Дата не может быть меньше сегодняшней", http.StatusBadRequest)
		return
	}

	result, err := DB.Exec(
		`INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`,

		task.Date, task.Title, task.Comment, task.Repeat,
	)
	if err != nil {
		utils.RespondError(w, "Ошибка базы данных: "+err.Error(), http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()
	utils.RespondJSON(w, map[string]string{"id": strconv.FormatInt(id, 10)})
}
