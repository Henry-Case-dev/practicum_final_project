package handlers

import (
	"net/http"
	"practicum_final_project/database"
	"practicum_final_project/utils"
)

// HandleDeleteTask обрабатывает удаление задачи
func HandleDeleteTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.RespondError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		utils.RespondError(w, "Не указан идентификатор", http.StatusBadRequest)
		return
	}

	_, err := database.DB.Exec(`DELETE FROM scheduler WHERE id = ?`, id)
	if err != nil {
		utils.RespondError(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	utils.RespondJSON(w, map[string]string{"result": "ok"}) // Не пустой ответ
}
