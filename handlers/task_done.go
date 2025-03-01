package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"practicum_final_project/models"
	"practicum_final_project/utils"
)

// HandleTaskDone обрабатывает отметку о выполнении задачи
func HandleTaskDone(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.RespondError(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		utils.RespondError(w, "Не указан идентификатор", http.StatusBadRequest)
		return
	}

	var task models.Task
	err := DB.QueryRow(
		`SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?`,
		id,
	).Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)

	if err == sql.ErrNoRows {
		utils.RespondError(w, "Задача не найдена", http.StatusNotFound)
		return
	}
	if err != nil {
		utils.RespondError(w, "Ошибка базы данных", http.StatusInternalServerError)
		return
	}

	if task.Repeat == "" {
		_, err = DB.Exec(`DELETE FROM scheduler WHERE id = ?`, id)
	} else {
		now := time.Now().UTC().Truncate(24 * time.Hour)
		nextDate, err := utils.NextDate(now, task.Date, task.Repeat)
		if err != nil {
			utils.RespondError(w, err.Error(), http.StatusBadRequest)
			return
		}
		_, err = DB.Exec(
			`UPDATE scheduler SET date = ? WHERE id = ?`,
			nextDate, id,
		)
	}

	if err != nil {
		utils.RespondError(w, "Ошибка базы данных: "+err.Error(), http.StatusInternalServerError)
		return
	}

	utils.RespondJSON(w, map[string]interface{}{})
}
