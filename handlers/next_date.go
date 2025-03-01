package handlers

import (
	"net/http"
	"practicum_final_project/utils"
)

// HandleNextDate обрабатывает запрос на вычисление следующей даты выполнения задачи.
func HandleNextDate(w http.ResponseWriter, r *http.Request) {
	// Предположим, что параметры передаются через URL: ?date=...&repeat=...
	date := r.URL.Query().Get("date")
	repeat := r.URL.Query().Get("repeat")
	if date == "" || repeat == "" {
		utils.RespondError(w, "Отсутствуют необходимые параметры", http.StatusBadRequest)
		return
	}
	now := utils.Now() // Если есть функция, возвращающая фиксированное текущее время для тестов.
	next, err := utils.NextDate(now, date, repeat)
	if err != nil {
		utils.RespondError(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(next))
}
