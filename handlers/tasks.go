package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"practicum_final_project/models"
	"practicum_final_project/utils"
)

// HandleTasks обрабатывает запрос на получение списка задач.
func HandleTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		// Если метод не GET, возвращаем ошибку
		utils.RespondError(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Получаем параметр поиска
	search := r.URL.Query().Get("search")
	limit := 50

	var rows *sql.Rows
	var err error

	if search != "" {
		// Если поиск можно интерпретировать как дату, пытаемся преобразовать ее
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
			// Иначе используем шаблон для поиска по заголовку и комментарию
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
		// Если поиск не задан, просто выбираем первые limit записей
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

	// Формируем срез задач
	tasks := make([]models.Task, 0)
	for rows.Next() {
		var task models.Task
		var id int64
		// Сканируем строку результата в переменные
		err := rows.Scan(&id, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			utils.RespondError(w, "Ошибка чтения данных", http.StatusInternalServerError)
			return
		}
		task.ID = strconv.FormatInt(id, 10)
		tasks = append(tasks, task)
	}

	// Возвращаем задачи в виде JSON
	utils.RespondJSON(w, map[string][]models.Task{"tasks": tasks})
}
