package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"practicum_final_project/models"
	"practicum_final_project/utils"
)

// HandleTask маршрутизирует входящие HTTP-запросы к обработчикам задач.
func HandleTask(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// Получение конкретной задачи по ID
		handleGetTask(w, r)
	case http.MethodPost:
		// Создание новой задачи
		handleCreateTask(w, r)
	case http.MethodPut:
		// Обновление уже существующей задачи
		handleUpdateTask(w, r)
	case http.MethodDelete:
		// Удаление задачи
		handleDeleteTask(w, r)
	default:
		// Метод не поддерживается – отправляем ошибку
		utils.RespondError(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	}
}

// handleGetTask получает задачу по переданному идентификатору.
func handleGetTask(w http.ResponseWriter, r *http.Request) {
	// Извлекаем параметр id из URL
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		utils.RespondError(w, "Не указан идентификатор", http.StatusBadRequest)
		return
	}

	var task models.Task
	// Выполняем запрос к базе данных для получения данных задачи
	err := DB.QueryRow(`SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?`, idStr).
		Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		// Если запись не найдена, отправляем соответствующую ошибку
		if err.Error() == "sql: no rows in result set" {
			utils.RespondError(w, "Задача не найдена", http.StatusNotFound)
			return
		}
		// Иначе ошибка базы данных
		utils.RespondError(w, "Ошибка базы данных", http.StatusInternalServerError)
		return
	}

	// Отправляем данные задачи в формате JSON
	utils.RespondJSON(w, task)
}

// handleCreateTask создает новую задачу.
func handleCreateTask(w http.ResponseWriter, r *http.Request) {
	var task models.Task
	// Декодируем тело запроса из JSON в структуру Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		utils.RespondError(w, "Неверный формат JSON", http.StatusBadRequest)
		return
	}

	// Проверяем обязательное поле Title
	if task.Title == "" {
		utils.RespondError(w, "Заголовок задачи обязателен", http.StatusBadRequest)
		return
	}

	// Определяем текущую дату с фиксированным часовым поясом
	now := utils.StartOfDay(time.Now())

	// Если дата не указана, используем сегодняшнюю
	if task.Date == "" {
		task.Date = now.Format("20060102")
	}

	// Парсим переданную дату
	parsedDate, err := utils.ParseDateString(task.Date)
	if err != nil {
		utils.RespondError(w, "Неверный формат даты", http.StatusBadRequest)
		return
	}

	// Если дата в прошлом – для повторяющихся задач вычисляем следующую, для разовых задаем сегодняшнюю
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

	// Сохраняем задачу в базе данных
	result, err := DB.Exec(
		`INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`,
		task.Date, task.Title, task.Comment, task.Repeat,
	)
	if err != nil {
		utils.RespondError(w, "Ошибка базы данных: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Получаем сгенерированный идентификатор задачи и отправляем его в ответе
	id, _ := result.LastInsertId()
	utils.RespondJSON(w, map[string]string{"id": strconv.FormatInt(id, 10)})
}

// handleUpdateTask обновляет существующую задачу.
func handleUpdateTask(w http.ResponseWriter, r *http.Request) {
	var task models.Task
	// Декодируем JSON из тела запроса
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

	// Преобразуем строковый идентификатор в число
	id, err := strconv.ParseInt(task.ID, 10, 64)
	if err != nil {
		utils.RespondError(w, "Некорректный идентификатор", http.StatusBadRequest)
		return
	}

	// Проверяем существование задачи в базе данных
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

	// Парсим переданную дату
	now := time.Now().UTC().Truncate(24 * time.Hour)
	parsedDate, err := time.Parse("20060102", task.Date)
	if err != nil {
		utils.RespondError(w, "Неверный формат даты", http.StatusBadRequest)
		return
	}

	// Если задача повторяющаяся – вычисляем новую дату, иначе, если дата меньше текущей, заменяем ее на текущую
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

	// Обновляем запись в базе данных
	_, err = DB.Exec(
		`UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?`,
		task.Date, task.Title, task.Comment, task.Repeat, id,
	)
	if err != nil {
		utils.RespondError(w, "Ошибка базы данных: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Отправляем успешный ответ
	utils.RespondJSON(w, map[string]interface{}{})
}

// handleDeleteTask удаляет задачу по идентификатору.
func handleDeleteTask(w http.ResponseWriter, r *http.Request) {
	// Извлекаем id из URL
	id := r.URL.Query().Get("id")
	if id == "" {
		utils.RespondError(w, "Не указан идентификатор", http.StatusBadRequest)
		return
	}

	// Выполняем операцию удаления в базе данных
	result, err := DB.Exec(`DELETE FROM scheduler WHERE id = ?`, id)
	if err != nil {
		utils.RespondError(w, "Ошибка базы данных: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Если запись не была найдена, отправляем ошибку
	affected, _ := result.RowsAffected()
	if affected == 0 {
		utils.RespondError(w, "Задача не найдена", http.StatusNotFound)
		return
	}

	// Отправляем пустой JSON-ответ
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
