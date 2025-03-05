package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"practicum_final_project/models"
	"practicum_final_project/utils"
)

// HandleCreateTask обрабатывает создание новой задачи.
// Если поле date отсутствует или пустое, то берётся значение параметра now из запроса (формат "20060102").
// При наличии правила повторения дата должна быть равна now, иначе возвращается ошибка.
func HandleCreateTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.RespondError(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var task models.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		utils.RespondError(w, "Неверный формат JSON", http.StatusBadRequest)
		return
	}
	if task.Title == "" {
		utils.RespondError(w, "Не указан заголовок задачи", http.StatusBadRequest)
		return
	}

	// Получаем now из запроса.
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
	log.Printf("DEBUG (CreateTask): now=%s", now.Format("20060102"))

	// Если поле date пустое – используем значение now.
	if task.Date == "" {
		task.Date = now.Format("20060102")
		log.Printf("DEBUG (CreateTask): date not provided, set to now=%s", task.Date)
	} else {
		if _, err := time.Parse("20060102", task.Date); err != nil {
			utils.RespondError(w, "Неверный формат даты", http.StatusBadRequest)
			return
		}
	}

	// При наличии правила повторения дата должна быть равна now.
	if task.Repeat != "" {
		if task.Date != now.Format("20060102") {
			log.Printf("DEBUG (CreateTask): repeat task but date (%s) != now (%s)", task.Date, now.Format("20060102"))
			utils.RespondError(w, "Дата должна быть сегодняшняя", http.StatusBadRequest)
			return
		}
	} else {
		// Если правило отсутствует и date меньше now – выдаём ошибку.
		parsedDate, _ := time.Parse("20060102", task.Date)
		if parsedDate.Before(now) {
			log.Printf("DEBUG (CreateTask): non-repeat task but date (%s) < now (%s)", task.Date, now.Format("20060102"))
			utils.RespondError(w, "Дата не может быть меньше сегодняшней", http.StatusBadRequest)
			return
		}
	}

	result, err := DB.Exec(
		`INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`,
		task.Date, task.Title, task.Comment, task.Repeat,
	)
	if err != nil {
		utils.RespondError(w, "Ошибка базы данных: "+err.Error(), http.StatusInternalServerError)
		return
	}
	id, err := result.LastInsertId()
	if err != nil {
		utils.RespondError(w, "Не удалось получить id задачи", http.StatusInternalServerError)
		return
	}
	log.Printf("DEBUG (CreateTask): task created with id=%d, date=%s", id, task.Date)

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	utils.RespondJSON(w, map[string]string{"id": strconv.FormatInt(id, 10)})
}
