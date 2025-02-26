package handlers

import (
	"database/sql"
	"net/http"
	"practicum_final_project/models"
	"practicum_final_project/utils"
	"time"
)

func HandleTaskDone(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.RespondError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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

	if task.Repeat == "" {
		_, err = db.Exec(`DELETE FROM scheduler WHERE id = ?`, id)
	} else {
		now := time.Now().UTC()
		nextDate, err := utils.NextDate(now, task.Date, task.Repeat)
		if err != nil {
			utils.RespondError(w, err.Error(), http.StatusBadRequest)
			return
		}
		_, err = db.Exec(
			`UPDATE scheduler 
            SET date = ? 
            WHERE id = ?`,
			nextDate, id,
		)
	}

	if err != nil {
		utils.RespondError(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	utils.RespondJSON(w, map[string]interface{}{})
}
