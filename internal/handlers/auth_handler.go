package handlers

import (
	"awesomeProject/internal/auth"
	"awesomeProject/internal/models"
	"awesomeProject/internal/services"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// AuthHandler структура для обработчиков аутентификации
type AuthHandler struct {
	entityService *services.EntityService
	secretKey     []byte
	refreshKey    []byte
}

// NewAuthHandler создает новый экземпляр AuthHandler
func NewAuthHandler(entityService *services.EntityService, secretKey, refreshKey []byte) *AuthHandler {
	return &AuthHandler{
		entityService: entityService,
		secretKey:     secretKey,
		refreshKey:    refreshKey,
	}
}

// RegisterHandler обрабатывает регистрацию нового пользователя
func (h *EntityHandler) RegisterHandler(w http.ResponseWriter, r *http.Request, entityType string) {
	var authEntity auth.Authenticatable

	switch entityType {
	case "users":
		authEntity = new(models.User)
	case "restaurants":
		authEntity = new(models.Restaurant)
	default:
		http.Error(w, "Invalid entity type", http.StatusBadRequest)
		return
	}

	log.Printf("Generating tokens for EntityTypeРег: %s", entityType)

	if err := json.NewDecoder(r.Body).Decode(authEntity); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	entityID, err := h.entityService.Register(r.Context(), authEntity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]string{"entityID": entityID}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		return
	}
}

func (h *AuthHandler) LoginHandler(w http.ResponseWriter, r *http.Request, entityType string) {
	var authEntity auth.Authenticatable
	switch entityType {
	case "users":
		authEntity = new(models.User)
	case "restaurants":
		authEntity = new(models.Restaurant)
	default:
		http.Error(w, "Invalid entity type", http.StatusBadRequest)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(authEntity); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Correctly log the type of the entity being authenticated
	secretKey, refreshTokenSecret := h.entityService.GetSecretKeys()
	accessToken, refreshToken, err := h.entityService.AuthenticateAndGenerateTokens(r.Context(), authEntity, secretKey, refreshTokenSecret)
	if err != nil {
		http.Error(w, "Failed to authenticate or generate tokens", http.StatusInternalServerError)
		return
	}

	setTokenCookies(w, accessToken, refreshToken)
	err = json.NewEncoder(w).Encode(map[string]string{"status": "success", "accessToken": accessToken, "refreshToken": refreshToken})
	if err != nil {
		return
	}
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
