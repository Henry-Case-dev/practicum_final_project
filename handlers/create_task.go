package handlers

import (
	"encoding/json"
	"net/http"
	"practicum_final_project/models"
	"practicum_final_project/utils"
	"strconv"
	"time"
)

// HandleCreateTask обрабатывает создание новой задачи
func HandleCreateTask(w http.ResponseWriter, r *http.Request) {
	var task models.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		utils.RespondError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if task.Title == "" {
		utils.RespondError(w, "Title is required", http.StatusBadRequest)
		return
	}

	now := time.Now().UTC()
	if task.Date == "" {
		task.Date = now.Format("20060102")
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

	res, err := db.Exec(
		`INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`,
		task.Date, task.Title, task.Comment, task.Repeat,
	)
	if err != nil {
		utils.RespondError(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	id, _ := res.LastInsertId()
	utils.RespondJSON(w, map[string]string{"id": strconv.FormatInt(id, 10)})
}
