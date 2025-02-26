package handlers

import (
	"database/sql"
	"net/http"
	"practicum_final_project/models"
	"practicum_final_project/utils"
	"strconv"
)

// HandleGetTask обрабатывает получение задачи по идентификатору
func HandleGetTask(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		utils.RespondError(w, "Не указан идентификатор", http.StatusBadRequest)
		return
	}

	var task models.Task
	row := db.QueryRow(
		`SELECT id, date, title, comment, repeat 
        FROM scheduler WHERE id = ?`,
		id,
	)

	var dbID int64
	err := row.Scan(&dbID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		if err == sql.ErrNoRows {
			utils.RespondError(w, "Задача не найдена", http.StatusNotFound)
		} else {
			utils.RespondError(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	task.ID = strconv.FormatInt(dbID, 10)
	utils.RespondJSON(w, task)
}
