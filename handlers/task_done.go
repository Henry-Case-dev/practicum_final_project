package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"practicum_final_project/models"
	"practicum_final_project/utils"
)

// HandleTaskDone обрабатывает отметку о выполнении задачи.
// Параметр now обязательный – время из запроса в формате 20060102.
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

	// Берём время из параметра запроса now.
	nowParam := r.URL.Query().Get("now")
	if nowParam == "" {
		utils.RespondError(w, "Отсутствует параметр now", http.StatusBadRequest)
		return
	}
	now, err := time.Parse("20060102", nowParam)
	if err != nil {
		utils.RespondError(w, "Неверный формат параметра now", http.StatusBadRequest)
		return
	}

	var task models.Task
	err = DB.QueryRow(
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

	// Если задача повторяется – вычисляем следующую дату относительно переданного now.
	if task.Repeat != "" {
		nextDate, err := utils.NextDate(now, task.Date, task.Repeat)
		if err != nil {
			utils.RespondError(w, err.Error(), http.StatusBadRequest)
			return
		}
		// Обновляем дату в БД.
		_, err = DB.Exec("UPDATE scheduler SET date = ? WHERE id = ?", nextDate, id)
		if err != nil {
			utils.RespondError(w, "Ошибка базы данных: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		// Для разовых задач – удаляем запись.
		_, err = DB.Exec("DELETE FROM scheduler WHERE id = ?", id)
		if err != nil {
			utils.RespondError(w, "Ошибка базы данных: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	utils.RespondJSON(w, map[string]interface{}{})
}
