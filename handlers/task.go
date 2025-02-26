package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"practicum_final_project/models"
	"practicum_final_project/utils"
	"strconv"
	"time"
)

// HandleTask обрабатывает запросы для задач
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
		utils.RespondError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGetTask обрабатывает получение задачи по идентификатору
func handleGetTask(w http.ResponseWriter, r *http.Request) {
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

// handleCreateTask обрабатывает создание новой задачи
func handleCreateTask(w http.ResponseWriter, r *http.Request) {
	var task models.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		utils.RespondError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if task.Title == "" {
		utils.RespondError(w, "Title is required", http.StatusBadRequest)
		return
	}

	now := time.Now().UTC()
	if task.Date == "" {
		task.Date = now.Format("20060102")
	}

	parsedDate, err := time.Parse("20060102", task.Date)
	if err != nil {
		utils.RespondError(w, "Invalid date format", http.StatusBadRequest)
		return
	}

	if task.Repeat != "" {
		next, err := utils.NextDate(now, task.Date, task.Repeat)
		if err != nil {
			utils.RespondError(w, err.Error(), http.StatusBadRequest)
			return
		}
		task.Date = next
	} else if parsedDate.Before(now) {
		task.Date = now.Format("20060102")
	}

	res, err := db.Exec(
		`INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`,
		task.Date, task.Title, task.Comment, task.Repeat,
	)
	if err != nil {
		utils.RespondError(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	id, _ := res.LastInsertId()
	utils.RespondJSON(w, map[string]string{"id": strconv.FormatInt(id, 10)})
}

// handleUpdateTask обрабатывает обновление задачи
func handleUpdateTask(w http.ResponseWriter, r *http.Request) {
	var task models.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		utils.RespondError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if task.ID == "" {
		utils.RespondError(w, "Не указан идентификатор", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(task.ID, 10, 64)
	if err != nil {
		utils.RespondError(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	var exists bool
	db.QueryRow("SELECT EXISTS(SELECT 1 FROM scheduler WHERE id = ?)", id).Scan(&exists)
	if !exists {
		utils.RespondError(w, "Задача не найдена", http.StatusNotFound)
		return
	}

	now := time.Now().UTC()
	if task.Title == "" {
		utils.RespondError(w, "Title is required", http.StatusBadRequest)
		return
	}

	parsedDate, err := time.Parse("20060102", task.Date)
	if err != nil {
		utils.RespondError(w, "Invalid date format", http.StatusBadRequest)
		return
	}

	if task.Repeat != "" {
		next, err := utils.NextDate(now, task.Date, task.Repeat)
		if err != nil {
			utils.RespondError(w, err.Error(), http.StatusBadRequest)
			return
		}
		task.Date = next
	} else if parsedDate.Before(now) {
		task.Date = now.Format("20060102")
	}

	_, err = db.Exec(
		`UPDATE scheduler 
        SET date = ?, title = ?, comment = ?, repeat = ? 
        WHERE id = ?`,
		task.Date, task.Title, task.Comment, task.Repeat, id,
	)

	if err != nil {
		utils.RespondError(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	utils.RespondJSON(w, map[string]interface{}{})
}

// handleDeleteTask обрабатывает удаление задачи
func handleDeleteTask(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		utils.RespondError(w, "Не указан идентификатор", http.StatusBadRequest)
		return
	}

	_, err := db.Exec(`DELETE FROM scheduler WHERE id = ?`, id)
	if err != nil {
		utils.RespondError(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	utils.RespondJSON(w, map[string]interface{}{})
}
