package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"practicum_final_project/models"
	"practicum_final_project/utils"
)

// HandleTask маршрутизирует запросы к обработчикам задач.
func HandleTask(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handleGetTask(w, r)
	case http.MethodPost:
		handleCreateTask(w, r)
	case http.MethodPut:
		handleUpdateTask(w, r)
	case http.MethodDelete:
		handleDeleteTask(w, r)
	default:
		utils.RespondError(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	}
}

// handleGetTask получает задачу по ID.
func handleGetTask(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		utils.RespondError(w, "Не указан идентификатор", http.StatusBadRequest)
		return
	}

	var task models.Task
	err := DB.QueryRow(`SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?`, idStr).
		Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			utils.RespondError(w, "Задача не найдена", http.StatusNotFound)
			return
		}
		utils.RespondError(w, "Ошибка базы данных", http.StatusInternalServerError)
		return
	}

	utils.RespondJSON(w, task)
}

// handleCreateTask создает новую задачу.
func handleCreateTask(w http.ResponseWriter, r *http.Request) {
	var task models.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		utils.RespondError(w, "Неверный формат JSON", http.StatusBadRequest)
		return
	}

	if task.Title == "" {
		utils.RespondError(w, "Заголовок задачи обязателен", http.StatusBadRequest)
		return
	}

	now := time.Now().UTC().Truncate(24 * time.Hour)
	// Если дата не указана, используем сегодняшнюю.
	if task.Date == "" {
		task.Date = now.Format("20060102")
	}

	parsedDate, err := time.Parse("20060102", task.Date)
	if err != nil {
		utils.RespondError(w, "Неверный формат даты", http.StatusBadRequest)
		return
	}

	if parsedDate.Before(now) {
		if task.Repeat != "" {
			nextDate, err := utils.NextDate(now, task.Date, task.Repeat)
			if err != nil {
				utils.RespondError(w, err.Error(), http.StatusBadRequest)
				return
			}
			task.Date = nextDate
		} else {
			task.Date = now.Format("20060102")
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

	id, _ := result.LastInsertId()
	utils.RespondJSON(w, map[string]string{"id": strconv.FormatInt(id, 10)})
}

// handleUpdateTask обновляет существующую задачу.
func handleUpdateTask(w http.ResponseWriter, r *http.Request) {
	var task models.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		utils.RespondError(w, "Неверный формат JSON", http.StatusBadRequest)
		return
	}

	if task.ID == "" {
		utils.RespondError(w, "Не указан идентификатор", http.StatusBadRequest)
		return
	}
	if task.Title == "" {
		utils.RespondError(w, "Заголовок задачи обязателен", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(task.ID, 10, 64)
	if err != nil {
		utils.RespondError(w, "Некорректный идентификатор", http.StatusBadRequest)
		return
	}

	var exists bool
	err = DB.QueryRow("SELECT EXISTS(SELECT 1 FROM scheduler WHERE id = ?)", id).Scan(&exists)
	if err != nil {
		utils.RespondError(w, "Ошибка базы данных", http.StatusInternalServerError)
		return
	}
	if !exists {
		utils.RespondError(w, "Задача не найдена", http.StatusNotFound)
		return
	}

	now := time.Now().UTC().Truncate(24 * time.Hour)
	parsedDate, err := time.Parse("20060102", task.Date)
	if err != nil {
		utils.RespondError(w, "Неверный формат даты", http.StatusBadRequest)
		return
	}

	if task.Repeat != "" {
		nextDate, err := utils.NextDate(now, task.Date, task.Repeat)
		if err != nil {
			utils.RespondError(w, err.Error(), http.StatusBadRequest)
			return
		}
		task.Date = nextDate
	} else if parsedDate.Before(now) {
		task.Date = now.Format("20060102")
	}

	_, err = DB.Exec(
		`UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?`,
		task.Date, task.Title, task.Comment, task.Repeat, id,
	)
	if err != nil {
		utils.RespondError(w, "Ошибка базы данных: "+err.Error(), http.StatusInternalServerError)
		return
	}

	utils.RespondJSON(w, map[string]interface{}{})
}

// handleDeleteTask удаляет задачу.
func handleDeleteTask(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		utils.RespondError(w, "Не указан идентификатор", http.StatusBadRequest)
		return
	}

	result, err := DB.Exec(`DELETE FROM scheduler WHERE id = ?`, id)
	if err != nil {
		utils.RespondError(w, "Ошибка базы данных: "+err.Error(), http.StatusInternalServerError)
		return
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		utils.RespondError(w, "Задача не найдена", http.StatusNotFound)
		return
	}

	utils.RespondJSON(w, map[string]interface{}{})
}

// RespondError отправляет JSON-ответ с сообщением об ошибке.
func RespondError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// RespondJSON отправляет JSON-ответ.
func RespondJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(data)
}
