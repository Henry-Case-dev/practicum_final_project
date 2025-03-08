package handlers

import (
	"net/http"
	"practicum_final_project/database"
	"practicum_final_project/utils"
)

// HandleDeleteTask обрабатывает удаление задачи по HTTP-запросу.
// Ожидается метод DELETE и параметр id в URL.
func HandleDeleteTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.RespondError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Извлекаем идентификатор задачи
	id := r.URL.Query().Get("id")
	if id == "" {
		utils.RespondError(w, "Не указан идентификатор", http.StatusBadRequest)
		return
	}

	// Выполняем удаление задачи в базе данных
	_, err := database.DB.Exec(`DELETE FROM scheduler WHERE id = ?`, id)
	if err != nil {
		utils.RespondError(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Отправляем пустой, но не пустой ответ (с ключом result)
	utils.RespondJSON(w, map[string]string{"result": "ok"})
}
