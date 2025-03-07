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

// HandleUpdateTask обновляет существующую задачу.
// В данной функции проверяется входной JSON, валидируются параметры (в том числе дата),
// затем производится обновление записи в базе данных.
func HandleUpdateTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		// Если метод не PUT, отправляем ошибку
		utils.RespondError(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Логируем сырой запрос для отладки
	log.Printf("DEBUG (UpdateTask): RawQuery=%s", r.URL.RawQuery)
	if err := r.ParseForm(); err == nil {
		log.Printf("DEBUG: URL.Query() = %+v", r.URL.Query())
		log.Printf("DEBUG (UpdateTask): Form=%v", r.Form)
	}

	var task models.Task
	// Декодирование JSON-тела запроса
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		utils.RespondError(w, "Неверный формат JSON", http.StatusBadRequest)
		return
	}

	// Проверяем обязательные поля
	if task.ID == "" {
		utils.RespondError(w, "Не указан идентификатор", http.StatusBadRequest)
		return
	}
	if task.Title == "" {
		utils.RespondError(w, "Заголовок задачи обязателен", http.StatusBadRequest)
		return
	}

	// Преобразуем id в числовое значение
	id, err := strconv.ParseInt(task.ID, 10, 64)
	if err != nil {
		utils.RespondError(w, "Некорректный идентификатор", http.StatusBadRequest)
		return
	}

	// Извлекаем параметр now из формы (ожидаемый формат "20060102")
	nowParam := r.FormValue("now")
	if nowParam == "" {
		log.Printf("DEBUG (UpdateTask): Отсутствует параметр now. RawQuery=%s, Form=%v", r.URL.RawQuery, r.Form)
		utils.RespondError(w, "Отсутствует параметр now", http.StatusBadRequest)
		return
	}
	now, err := time.Parse("20060102", nowParam)
	if err != nil {
		log.Printf("DEBUG (UpdateTask): Неверный формат параметра now. RawQuery=%s, Form=%v", r.URL.RawQuery, r.Form)
		utils.RespondError(w, "Неверный формат параметра now", http.StatusBadRequest)
		return
	}
	log.Printf("DEBUG (UpdateTask): now=%s", now.Format("20060102"))

	// Если дата не указана – используем значение параметра now
	if task.Date == "" {
		task.Date = now.Format("20060102")
		log.Printf("DEBUG (UpdateTask): date not provided, set to now=%s", task.Date)
	} else {
		// Проверяем формат даты
		if _, err := time.Parse("20060102", task.Date); err != nil {
			utils.RespondError(w, "Неверный формат даты", http.StatusBadRequest)
			return
		}
	}

	// Валидация даты в зависимости от типа задачи (повторяющаяся или нет)
	if task.Repeat != "" {
		// Для повторяющихся задач дата должна быть равна now
		if task.Date != now.Format("20060102") {
			log.Printf("DEBUG (UpdateTask): repeat task but date (%s) != now (%s)", task.Date, now.Format("20060102"))
			utils.RespondError(w, "Дата должна быть сегодняшняя", http.StatusBadRequest)
			return
		}
	} else {
		// Для не повторяющихся задач, если указанная дата меньше now, заменяем её
		parsedDate, _ := time.Parse("20060102", task.Date)
		if parsedDate.Before(now) {
			log.Printf("DEBUG (UpdateTask): non-repeat task, date (%s) is before now (%s); replacing with now", task.Date, now.Format("20060102"))
			task.Date = now.Format("20060102")
		}
	}

	// Проверяем, существует ли задача с данным идентификатором
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

	// Обновляем запись в базе данных с новыми значениями
	_, err = DB.Exec(
		`UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?`,
		task.Date, task.Title, task.Comment, task.Repeat, id,
	)
	if err != nil {
		utils.RespondError(w, "Ошибка базы данных: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Логируем успешное обновление задачи
	log.Printf("DEBUG (UpdateTask): task id=%d updated successfully, new date=%s", id, task.Date)
	utils.RespondJSON(w, map[string]interface{}{})
}
