package handlers

import (
	"encoding/json"
	"net/http"

	"awesomeProject/internal/models"
	"awesomeProject/internal/services"
)

// AuthHandler структура для обработчиков аутентификации
type AuthHandler struct {
	userService *services.UserService
}

// NewAuthHandler создает новый экземпляр AuthHandler
func NewAuthHandler(userService *services.UserService) *AuthHandler {
	return &AuthHandler{
		userService: userService,
	}
}

// RegisterHandler обрабатывает регистрацию нового пользователя
func (h *AuthHandler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
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
	json.NewEncoder(w).Encode(response)
}

// LoginHandler обрабатывает вход пользователя
func (h *AuthHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var creds models.UserCredentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	token, err := h.userService.Authenticate(r.Context(), creds.Email, creds.Password)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	response := map[string]string{"token": token}
	json.NewEncoder(w).Encode(response)
}
