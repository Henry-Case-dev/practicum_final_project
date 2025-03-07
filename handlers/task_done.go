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
// В данной функции используется текущее время (currentTime), и если задача повторяется – вычисляется следующая дата;
// для разовых задач запись удаляется.
func HandleTaskDone(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		// Если метод не POST, возвращаем ошибку
		utils.RespondError(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Логируем содержимое запроса для отладки
	log.Printf("DEBUG (TaskDone): RawQuery=%s", r.URL.RawQuery)
	if err := r.ParseForm(); err == nil {
		log.Printf("DEBUG (TaskDone): Form=%v", r.Form)
	}

	// Извлекаем идентификатор задачи
	id := r.FormValue("id")
	if id == "" {
		utils.RespondError(w, "Не указан идентификатор", http.StatusBadRequest)
		return
	}

	// Устанавливаем текущее время, обрезая его до полуночи
	currentTime := time.Now().Local().Truncate(24 * time.Hour)
	log.Printf("DEBUG (TaskDone): currentTime=%s", currentTime.Format("20060102"))

	var task models.Task
	// Получаем данные задачи из базы данных
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

	// Если задача повторяется – вычисляем следующую дату и обновляем запись
	if task.Repeat != "" {
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
		// Для разовых задач удаляем запись из базы
		_, err = DB.Exec("DELETE FROM scheduler WHERE id = ?", id)
		if err != nil {
			utils.RespondError(w, "Ошибка базы данных: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Отправляем успешный ответ
	utils.RespondJSON(w, map[string]interface{}{})
}
