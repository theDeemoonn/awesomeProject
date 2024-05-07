package handlers

import (
	"awesomeProject/internal/models"
	"awesomeProject/internal/services"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// AuthHandler структура для обработчиков аутентификации
type AuthHandler struct {
	userService *services.UserService
	secretKey   []byte
	refreshKey  []byte
}

// NewAuthHandler создает новый экземпляр AuthHandler
func NewAuthHandler(userService *services.UserService, secretKey, refreshKey []byte) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		secretKey:   secretKey,
		refreshKey:  refreshKey,
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

func (h *AuthHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var creds models.UserCredentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.userService.Authenticate(r.Context(), creds.Email, creds.Password)
	if err != nil {
		log.Printf("Authentication failed for email %s with error: %v", creds.Email, err)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}
	if err != nil {
		http.Error(w, "Invalid credentialsLogin", http.StatusUnauthorized)
		return
	}

	secretKey, refreshTokenSecret := h.userService.GetSecretKeys() // Предполагаем, что эти ключи получены надлежащим образом
	accessToken, refreshToken, err := h.userService.GenerateAndStoreToken(r.Context(), user, secretKey, refreshTokenSecret)
	if err != nil {
		http.Error(w, "Failed to generate tokens", http.StatusInternalServerError)
		return
	}

	// Установка куки с токенами
	setTokenCookies(w, accessToken, refreshToken)

	// Отправка ответа, что аутентификация прошла успешно
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func setTokenCookies(w http.ResponseWriter, accessToken, refreshToken string) {
	// Установка куки для доступного токена
	http.SetCookie(w, &http.Cookie{
		Name:     "AccessToken",
		Value:    accessToken,
		Expires:  time.Now().Add(15 * time.Minute), // срок действия доступного токена
		HttpOnly: true,                             // защита от доступа через JavaScript
		Secure:   true,                             // куки отправляются только по HTTPS
		Path:     "/",
		SameSite: http.SameSiteStrictMode, // предотвращение отправки куки вместе с кросс-сайтовыми запросами
	})

	// Установка куки для рефреш токена
	http.SetCookie(w, &http.Cookie{
		Name:     "RefreshToken",
		Value:    refreshToken,
		Expires:  time.Now().Add(7 * 24 * time.Hour), // срок действия рефреш токена
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
	})
}
