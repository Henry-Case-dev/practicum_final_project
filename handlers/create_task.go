package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"practicum_final_project/models"
	"practicum_final_project/utils"
)

// HandleCreateTask обрабатывает создание новой задачи.
// Если поле date пустое – вместо него берётся значение параметра now из запроса (в формате 20060102).
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

	// Получаем now из запроса, он обязателен для обработки повторяющихся задач.
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

	// Если date не указана или пуста – используем значение now.
	if task.Date == "" {
		task.Date = now.Format("20060102")
	} else {
		if _, err := time.Parse("20060102", task.Date); err != nil {
			utils.RespondError(w, "Неверный формат даты", http.StatusBadRequest)
			return
		}
	}

	// При наличии правила повторения, дата должна быть равна now.
	if task.Repeat != "" {
		if task.Date != now.Format("20060102") {
			utils.RespondError(w, "Дата должна быть сегодняшняя", http.StatusBadRequest)
			return
		}
	} else {
		// Если правило отсутствует и дата меньше now, выдаём ошибку.
		parsedDate, _ := time.Parse("20060102", task.Date)
		if parsedDate.Before(now) {
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
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	utils.RespondJSON(w, map[string]string{"id": strconv.FormatInt(id, 10)})
}
