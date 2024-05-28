package handlers

import (
	"awesomeProject/internal/auth"
	"context"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"net/http"
	"os"
	"strings"

	"awesomeProject/internal/models"
	"awesomeProject/internal/services"
)

// EntityHandler структура для обработчиков пользователей
type EntityHandler struct {
	entityService *services.EntityService
	redisService  *services.RedisService
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// NewEntityHandler создает новый экземпляр EntityHandler
func NewEntityHandler(userService *services.EntityService, redisService *services.RedisService) *EntityHandler {
	return &EntityHandler{
		entityService: userService,
		redisService:  redisService,
	}
}

// LoginEntityHandler обрабатывает аутентификацию пользователя
func (h *EntityHandler) LoginEntityHandler(w http.ResponseWriter, r *http.Request, entityType string) {
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
	if err := json.NewDecoder(r.Body).Decode(&authEntity); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	token, err := h.entityService.Authenticate(r.Context(), authEntity.GetEmail(), authEntity.GetPassword(), authEntity)
	if err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(token)

}

// GetEntityById обрабатывает получение данных сущности по ID
func (h *EntityHandler) GetEntityById(w http.ResponseWriter, r *http.Request, entityType string) {
	vars := mux.Vars(r)
	entityIDStr := vars["id"]

	entityID, err := primitive.ObjectIDFromHex(entityIDStr)
	if err != nil {
		http.Error(w, "Invalid entity ID", http.StatusBadRequest)
		return
	}

	var entity interface{}
	switch entityType {
	case "users":
		entity = new(models.User)
	case "restaurants":
		entity = new(models.Restaurant)
	default:
		http.Error(w, "Invalid entity type", http.StatusBadRequest)
		return
	}

	err = h.redisService.GetCachedEntity(entityID.Hex(), &entity)
	if err == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(entity)
		return
	}

	if err != redis.Nil {
		http.Error(w, "Error getting "+entityType+" from cache: "+err.Error(), http.StatusInternalServerError)
		return
	}

	entity, err = h.entityService.GetEntity(r.Context(), entityID.Hex(), entityType)
	if err != nil || entity == nil {
		http.Error(w, entityType+" not found", http.StatusNotFound)
		return
	}

	err = h.redisService.CacheEntity(entityIDStr, entity)
	if err != nil {
		http.Error(w, "Failed to cache "+entityType, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entity)
}

// UpdateEntity обрабатывает обновление данных сущности
func (h *EntityHandler) UpdateEntity(w http.ResponseWriter, r *http.Request, entityType string) {
	entityID := r.URL.Query().Get("id")

	var entity interface{}
	switch entityType {
	case "users":
		entity = new(models.User)
	case "restaurants":
		entity = new(models.Restaurant)
	default:
		http.Error(w, "Invalid entity type", http.StatusBadRequest)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&entity); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.entityService.UpdateEntity(r.Context(), entityID, entity, entityType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]string{"message": entityType + " updated successfully"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ChangePasswordHandler обрабатывает изменение пароля сущности
func (h *EntityHandler) ChangePasswordHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	entityIDStr := vars["id"]

	entityID, err := primitive.ObjectIDFromHex(entityIDStr)
	if err != nil {
		http.Error(w, "Invalid entity ID", http.StatusBadRequest)
		return
	}

	var changePasswordRequest ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&changePasswordRequest); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	entityType := r.URL.Query().Get("type")
	var entity auth.Authenticatable

	switch entityType {
	case "users":
		entity = new(models.User)
	case "restaurants":
		entity = new(models.Restaurant)
	default:
		http.Error(w, "Invalid entity type", http.StatusBadRequest)
		return
	}

	err = h.entityService.ChangePassword(r.Context(), entityID.Hex(), changePasswordRequest.OldPassword, changePasswordRequest.NewPassword, entity)
	if err != nil {
		http.Error(w, "Failed to change password: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "password changed successfully"})
}

// DeleteEntity обрабатывает удаление сущности
func (h *EntityHandler) DeleteEntity(w http.ResponseWriter, r *http.Request, entityType string) {
	entityID := r.URL.Query().Get("id")

	err := h.entityService.DeleteEntity(r.Context(), entityID, entityType)
	if err != nil {
		http.Error(w, entityType+" not found or delete failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]string{"message": entityType + " deleted successfully"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetAllUsers обрабатывает получение всех пользователей
func (h *EntityHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.entityService.GetAllUsers(r.Context())
	if err != nil {
		http.Error(w, "Error getting users", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// GetAllRestaurants обрабатывает получение всех ресторанов
func (h *EntityHandler) GetAllRestaurants(w http.ResponseWriter, r *http.Request) {
	restaurants, err := h.entityService.GetAllRestaurants(r.Context())
	if err != nil {
		http.Error(w, "Error getting restaurants", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(restaurants)

}

// GetEntity обрабатывает получение данных сущности
func (h *EntityHandler) GetEntity(w http.ResponseWriter, r *http.Request) {
	tokenString := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if tokenString == "" {
		http.Error(w, "Authorization header is required", http.StatusUnauthorized)
		return
	}

	claims, err := auth.ValidateToken(tokenString, os.Getenv("SECRET_KEY"))
	if err != nil {
		http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
		return
	}

	entity, err := h.entityService.GetEntity(r.Context(), claims.UserID.Hex(), claims.EntityType)
	if err != nil {
		log.Printf("Error finding entity: %v", err)
		http.Error(w, claims.EntityType+" not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entity)
}

// AddFavoriteRestaurantHandler обрабатывает добавление ресторана в список избранных
func (h *EntityHandler) AddFavoriteRestaurantHandler(w http.ResponseWriter, r *http.Request) {
	// Извлечение ID пользователя из контекста запроса
	userID, err := getUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Извлечение ID ресторана из запроса
	vars := mux.Vars(r)
	restaurantIDHex := vars["restaurant_id"]
	restaurantID, err := primitive.ObjectIDFromHex(restaurantIDHex)
	if err != nil {
		http.Error(w, "Invalid restaurant ID", http.StatusBadRequest)
		return
	}

	// Добавление ресторана в список избранных
	err = h.entityService.AddFavoriteRestaurant(r.Context(), userID, restaurantID)
	if err != nil {
		http.Error(w, "Failed to add favorite restaurant", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Restaurant added to favorites"))
}

// GetFavoriteRestaurantsHandler обрабатывает получение списка избранных ресторанов пользователя
func (h *EntityHandler) GetFavoriteRestaurantsHandler(w http.ResponseWriter, r *http.Request) {
	// Извлечение ID пользователя из контекста запроса
	userID, err := getUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Получение списка избранных ресторанов
	favoriteRestaurants, err := h.entityService.GetFavoriteRestaurants(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to get favorite restaurants", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(favoriteRestaurants)
}

// getUserIDFromContext извлекает ID пользователя из контекста
func getUserIDFromContext(ctx context.Context) (primitive.ObjectID, error) {
	userClaims, ok := ctx.Value("userClaims").(*auth.JWTClaims)
	if !ok || userClaims == nil {
		return primitive.NilObjectID, errors.New("no user claims found in context")
	}

	userID, err := primitive.ObjectIDFromHex(userClaims.UserID.Hex())
	if err != nil {
		return primitive.NilObjectID, errors.New("invalid user ID")
	}

	return userID, nil
}
