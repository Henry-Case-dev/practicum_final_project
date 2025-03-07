package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"practicum_final_project/models"
	"practicum_final_project/utils"
)

// HandleTaskDone обрабатывает отметку о выполнении задачи.
// Параметр now больше не берётся из запроса – используется текущее время.
// Если задача повторяется, вычисляется следующая дата (на основе currentTime + правило);
// для разовых задач запись удаляется.
func HandleTaskDone(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.RespondError(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Логируем содержимое запроса.
	log.Printf("DEBUG (TaskDone): RawQuery=%s", r.URL.RawQuery)
	if err := r.ParseForm(); err == nil {
		log.Printf("DEBUG (TaskDone): Form=%v", r.Form)
	}

	id := r.FormValue("id")
	if id == "" {
		utils.RespondError(w, "Не указан идентификатор", http.StatusBadRequest)
		return
	}

	// Используем текущее локальное время, обрезанное до начала дня, как currentTime.
	currentTime := time.Now().Local().Truncate(24 * time.Hour)
	log.Printf("DEBUG (TaskDone): currentTime=%s", currentTime.Format("20060102"))

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
	log.Printf("DEBUG (TaskDone): task retrieved, id=%s, date=%s, repeat=%s", id, task.Date, task.Repeat)

	if task.Repeat != "" {
		// Вычисляем следующую дату относительно currentTime.
		nextDate, err := utils.NextDate(currentTime, task.Date, task.Repeat)
		if err != nil {
			utils.RespondError(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Printf("DEBUG (TaskDone): updating task id=%s, current date=%s, next date=%s", id, task.Date, nextDate)
		_, err = DB.Exec("UPDATE scheduler SET date = ? WHERE id = ?", nextDate, id)
		if err != nil {
			utils.RespondError(w, "Ошибка базы данных: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		// Для разовых задач — удаляем запись.
		_, err = DB.Exec("DELETE FROM scheduler WHERE id = ?", id)
		if err != nil {
			utils.RespondError(w, "Ошибка базы данных: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	utils.RespondJSON(w, map[string]interface{}{})
}
