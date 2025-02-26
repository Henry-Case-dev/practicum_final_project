package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"practicum_final_project/models"
	"practicum_final_project/utils"
)

// HandleUpdateTask обрабатывает обновление задачи
func HandleUpdateTask(w http.ResponseWriter, r *http.Request) {
	var task models.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		utils.RespondError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if task.ID == "" {
		utils.RespondError(w, "Не указан идентификатор", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(task.ID, 10, 64)
	if err != nil {
		utils.RespondError(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	var exists bool
	db.QueryRow("SELECT EXISTS(SELECT 1 FROM scheduler WHERE id = ?)", id).Scan(&exists)
	if !exists {
		utils.RespondError(w, "Задача не найдена", http.StatusNotFound)
		return
	}

	now := time.Now().UTC()
	if task.Title == "" {
		utils.RespondError(w, "Title is required", http.StatusBadRequest)
		return
	}

	parsedDate, err := time.Parse("20060102", task.Date)
	if err != nil {
		utils.RespondError(w, "Invalid date format", http.StatusBadRequest)
		return
	}

	if task.Repeat != "" {
		next, err := utils.NextDate(now, task.Date, task.Repeat)
		if err != nil {
			utils.RespondError(w, err.Error(), http.StatusBadRequest)
			return
		}
		task.Date = next
	} else if parsedDate.Before(now) {
		task.Date = now.Format("20060102")
	}

	_, err = db.Exec(
		`UPDATE scheduler 
        SET date = ?, title = ?, comment = ?, repeat = ? 
        WHERE id = ?`,
		task.Date, task.Title, task.Comment, task.Repeat, id,
	)

	if err != nil {
		utils.RespondError(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	utils.RespondJSON(w, map[string]interface{}{})
}
