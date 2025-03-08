package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

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

	// В зависимости от наличия параметра search выполняем разные запросы
	if search == "" {
		// Если параметр поиска не указан, получаем все задачи, отсортированные по дате
		rows, err = DB.Query(`
            SELECT id, date, title, comment, repeat 
            FROM scheduler 
            ORDER BY date ASC
            LIMIT ?
        `, limit)
	} else {
		// Если параметр поиска указан, фильтруем задачи по полям title и comment
		search = "%" + search + "%"
		rows, err = DB.Query(`
            SELECT id, date, title, comment, repeat 
            FROM scheduler 
            WHERE title LIKE ? OR comment LIKE ?
            ORDER BY date ASC
            LIMIT ?
        `, search, search, limit)
	}

	// Проверяем наличие ошибки при выполнении запроса
	if err != nil {
		utils.RespondError(w, "Ошибка базы данных", http.StatusInternalServerError)
		return
	}
	// Отложенное закрытие ресурса
	defer rows.Close()

	// Формируем срез задач
	tasks := make([]models.Task, 0)
	for rows.Next() {
		var task models.Task
		var id int64
		// Сканируем строку результата в переменные
		err := rows.Scan(&id, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			log.Printf("Ошибка сканирования строки: %v", err)
			utils.RespondError(w, "Ошибка при получении данных", http.StatusInternalServerError)
			return
		}
		task.ID = strconv.FormatInt(id, 10)
		tasks = append(tasks, task)
	}

	// Проверяем, нет ли ошибки при обработке итераций
	if err = rows.Err(); err != nil {
		log.Printf("Ошибка при итерации по результатам: %v", err)
		utils.RespondError(w, "Ошибка при получении данных", http.StatusInternalServerError)
		return
	}

	// Возвращаем задачи в виде JSON
	utils.RespondJSON(w, map[string][]models.Task{"tasks": tasks})
}
