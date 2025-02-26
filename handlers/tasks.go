package handlers

import (
	"database/sql"
	"net/http"
	"practicum_final_project/models"
	"practicum_final_project/utils"
	"strconv"
	"time"
)

// HandleTasks обрабатывает получение списка задач
func HandleTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.RespondError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	search := r.URL.Query().Get("search")
	limit := 50

	var rows *sql.Rows
	var err error

	if search != "" {
		if date, err := time.Parse("02.01.2006", search); err == nil {
			rows, err = db.Query(
				`SELECT id, date, title, comment, repeat 
                FROM scheduler 
                WHERE date = ? 
                ORDER BY date 
                LIMIT ?`,
				date.Format("20060102"),
				limit,
			)
		} else {
			searchPattern := "%" + search + "%"
			rows, err = db.Query(
				`SELECT id, date, title, comment, repeat 
                FROM scheduler 
                WHERE title LIKE ? OR comment LIKE ? 
                ORDER BY date 
                LIMIT ?`,
				searchPattern,
				searchPattern,
				limit,
			)
		}
	} else {
		rows, err = db.Query(
			`SELECT id, date, title, comment, repeat 
            FROM scheduler 
            ORDER BY date 
            LIMIT ?`,
			limit,
		)
	}

	if err != nil {
		utils.RespondError(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	tasks := make([]models.Task, 0)
	for rows.Next() {
		var id int64
		var t models.Task
		err := rows.Scan(&id, &t.Date, &t.Title, &t.Comment, &t.Repeat)
		if err != nil {
			utils.RespondError(w, "Error reading tasks", http.StatusInternalServerError)
			return
		}
		t.ID = strconv.FormatInt(id, 10)
		tasks = append(tasks, t)
	}

	if tasks == nil {
		tasks = make([]models.Task, 0)
	}

	utils.RespondJSON(w, map[string][]models.Task{"tasks": tasks})
}
