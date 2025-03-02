package handlers

import (
	"net/http"
	"time"

	"practicum_final_project/utils"
)

// HandleNextDate обрабатывает запрос на вычисление следующей даты выполнения задачи.
func HandleNextDate(w http.ResponseWriter, r *http.Request) {
	// Ожидаются параметры: now, date и repeat.
	nowParam := r.URL.Query().Get("now")
	date := r.URL.Query().Get("date")
	repeat := r.URL.Query().Get("repeat")
	if nowParam == "" || date == "" || repeat == "" {
		utils.RespondError(w, "Отсутствуют необходимые параметры", http.StatusBadRequest)
		return
	}
	// Разбор параметра now в формате 20060102
	now, err := time.Parse("20060102", nowParam)
	if err != nil {
		utils.RespondError(w, "Неверный формат параметра now", http.StatusBadRequest)
		return
	}
	// Вычисляем следующую дату относительно указанного now
	next, err := utils.NextDate(now, date, repeat)
	if err != nil {
		utils.RespondError(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(next))
}
