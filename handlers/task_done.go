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
// Параметр now обязателен – время из запроса (формат "20060102").
// Если задача повторяется, вычисляется следующая дата (на основе now + правило);
// для разовых задач запись удаляется.
func HandleTaskDone(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.RespondError(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Логируем полное содержимое запроса.
	log.Printf("DEBUG (TaskDone): RawQuery=%s", r.URL.RawQuery)
	if err := r.ParseForm(); err == nil {
		log.Printf("DEBUG (TaskDone): Form=%v", r.Form)
	}

	id := r.FormValue("id")
	if id == "" {
		utils.RespondError(w, "Не указан идентификатор", http.StatusBadRequest)
		return
	}

	// Получаем параметр now через FormValue.
	nowParam := r.FormValue("now")
	if nowParam == "" {
		log.Printf("DEBUG (TaskDone): Отсутствует параметр now. RawQuery=%s, Form=%v", r.URL.RawQuery, r.Form)
		utils.RespondError(w, "Отсутствует параметр now", http.StatusBadRequest)
		return
	}
	log.Printf("DEBUG (TaskDone): received nowParam=%s", nowParam)
	now, err := time.Parse("20060102", nowParam)
	if err != nil {
		log.Printf("DEBUG (TaskDone): Неверный формат параметра now. RawQuery=%s, Form=%v", r.URL.RawQuery, r.Form)
		utils.RespondError(w, "Неверный формат параметра now", http.StatusBadRequest)
		return
	}
	log.Printf("DEBUG (TaskDone): now=%s", now.Format("20060102"))

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
	log.Printf("DEBUG (TaskDone): task retrieved, id=%s, date=%s, repeat=%s", id, task.Date, task.Repeat)

	// Если задача повторяется – проверяем, что дата из БД равна now.
	if task.Repeat != "" {
		if task.Date != now.Format("20060102") {
			log.Printf("DEBUG (TaskDone): повторяющаяся задача, но date (%s) не равна now (%s)", task.Date, now.Format("20060102"))
			utils.RespondError(w, "Дата должна быть сегодняшняя", http.StatusBadRequest)
			return
		}
		// Вычисляем следующую дату относительно now.
		nextDate, err := utils.NextDate(now, task.Date, task.Repeat)
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
