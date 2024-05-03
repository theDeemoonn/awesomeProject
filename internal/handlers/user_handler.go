package handlers

import (
	"encoding/json"
	"net/http"

	"awesomeProject/internal/models"
	"awesomeProject/internal/services"
)

// UserHandler структура для обработчиков пользователей
type UserHandler struct {
	userService *services.UserService
}

// NewUserHandler создает новый экземпляр UserHandler
func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// RegisterUser обрабатывает регистрацию нового пользователя
func (h *UserHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID, err := h.userService.Register(r.Context(), &user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]string{"userID": userID}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetUser обрабатывает получение данных пользователя
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	// Это пример того, как можно извлечь параметр из запроса, например, ID пользователя
	userID := r.URL.Query().Get("id")

	user, err := h.userService.GetUser(r.Context(), userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
