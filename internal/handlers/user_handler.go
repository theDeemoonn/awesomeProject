package handlers

import (
	"awesomeProject/internal/auth"
	"encoding/json"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"strings"

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

// GetUserById обрабатывает получение данных пользователя
func (h *UserHandler) GetUserById(w http.ResponseWriter, r *http.Request) {
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

// UpdateUser обрабатывает обновление данных пользователя
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	// Это пример того, как можно извлечь параметр из запроса, например, ID пользователя
	userID := r.URL.Query().Get("id")

	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.userService.UpdateUser(r.Context(), userID, &user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]string{"message": "User updated successfully"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeleteUser обрабатывает удаление пользователя
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	// Это пример того, как можно извлечь параметр из запроса, например, ID пользователя
	userID := r.URL.Query().Get("id")

	err := h.userService.DeleteUser(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]string{"message": "User deleted successfully"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetUserHandler обрабатывает получение данных пользователя

// GetAllUsers обрабатывает получение всех пользователей
func (h *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userService.GetAllUsers(r.Context())
	if err != nil {
		http.Error(w, "Error getting users", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// GetUser обрабатывает получение данных пользователя
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	if err := godotenv.Load("/Users/dima/go/src/awesomeProject/.env"); err != nil {
		log.Fatal("Error loading .env file")
	}

	secretKey := os.Getenv("SECRET_KEY")
	if secretKey == "" {
		log.Fatal("JWT_SECRET_KEY is not set in environment variables")
	}
	// Извлекаем токен из заголовка Authorization
	tokenString := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if tokenString == "" {
		http.Error(w, "Authorization header is required", http.StatusUnauthorized)
		return
	}

	// Валидация и разбор токена
	claims, err := auth.ValidateToken(tokenString, secretKey)
	if err != nil {
		http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
		return
	}

	// Проверка, что токен не просрочен и подпись корректна
	//if !token.Valid {
	//	http.Error(w, "Invalid token", http.StatusUnauthorized)
	//	return
	//}

	// Извлечение ID пользователя из токена
	userID := claims.UserID.Hex()

	// Запрос информации о пользователе из базы данных
	user, err := h.userService.GetUser(r.Context(), userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Отправка данных пользователя клиенту
	response, err := json.Marshal(user)
	if err != nil {
		http.Error(w, "Failed to serialize user data", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}
