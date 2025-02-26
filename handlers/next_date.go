package handlers

import (
	"log"
	"net/http"
	"practicum_final_project/utils"
	"time"
)

// HandleNextDate обрабатывает получение следующей даты выполнения задачи
func HandleNextDate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.RespondError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	date := r.URL.Query().Get("date")
	repeat := r.URL.Query().Get("repeat")

	if date == "" || repeat == "" {
		utils.RespondError(w, "Missing date or repeat parameter", http.StatusBadRequest)
		return
	}

	now := time.Now().UTC()
	log.Printf("HandleNextDate called with date: %s, repeat: %s", date, repeat)
	nextDate, err := utils.NextDate(now, date, repeat)
	if err != nil {
		utils.RespondError(w, err.Error(), http.StatusBadRequest)
		return
	}

	utils.RespondJSON(w, map[string]string{"date": nextDate})
}
