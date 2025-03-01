package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"practicum_final_project/models"
	"practicum_final_project/utils"
)

// HandleTasks обрабатывает получение списка задач
func HandleTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.RespondError(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	search := r.URL.Query().Get("search")
	limit := 50

	var rows *sql.Rows
	var err error

	if search != "" {
		if date, err := time.Parse("02.01.2006", search); err == nil {
			search = date.Format("20060102")
			rows, err = DB.Query(
				`SELECT id, date, title, comment, repeat 
                FROM scheduler 
                WHERE date = ? 
                ORDER BY date LIMIT ?`,
				search, limit,
			)
		} else {
			searchPattern := "%" + search + "%"
			rows, err = DB.Query(
				`SELECT id, date, title, comment, repeat 
                FROM scheduler 
                WHERE title LIKE ? OR comment LIKE ? 
                ORDER BY date LIMIT ?`,
				searchPattern, searchPattern, limit,
			)
		}
	} else {
		rows, err = DB.Query(
			`SELECT id, date, title, comment, repeat 
            FROM scheduler 
            ORDER BY date LIMIT ?`,
			limit,
		)
	}

	if err != nil {
		utils.RespondError(w, "Ошибка базы данных", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	tasks := make([]models.Task, 0)
	for rows.Next() {
		var task models.Task
		var id int64
		err := rows.Scan(&id, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			utils.RespondError(w, "Ошибка чтения данных", http.StatusInternalServerError)
			return
		}
		task.ID = strconv.FormatInt(id, 10)
		tasks = append(tasks, task)
	}

	utils.RespondJSON(w, map[string][]models.Task{"tasks": tasks})
}
