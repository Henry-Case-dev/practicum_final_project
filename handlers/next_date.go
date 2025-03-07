package handlers

import (
	"net/http"
	"time"

	"practicum_final_project/utils"
)

// HandleNextDate обрабатывает запрос на вычисление следующей даты выполнения задачи.
// Ожидаются параметры: now, date и repeat.
func HandleNextDate(w http.ResponseWriter, r *http.Request) {
	// Извлекаем параметры запроса
	nowParam := r.URL.Query().Get("now")
	date := r.URL.Query().Get("date")
	repeat := r.URL.Query().Get("repeat")
	if nowParam == "" || date == "" || repeat == "" {
		utils.RespondError(w, "Отсутствуют необходимые параметры", http.StatusBadRequest)
		return
	}
	// Парсим параметр now согласно формату "20060102"
	now, err := time.Parse("20060102", nowParam)
	if err != nil {
		utils.RespondError(w, "Неверный формат параметра now", http.StatusBadRequest)
		return
	}
	// Вычисляем следующую дату с использованием функции NextDate
	next, err := utils.NextDate(now, date, repeat)
	if err != nil {
		utils.RespondError(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Отправляем результат в виде текста
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(next))
}
