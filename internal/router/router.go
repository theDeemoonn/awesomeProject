package router

import (
	"awesomeProject/internal/auth"
	"awesomeProject/internal/handlers"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
)

// InitializeRouter настраивает и возвращает роутер
func InitializeRouter(userHandler *handlers.EntityHandler, authHandler *handlers.AuthHandler, restaurantHandler *handlers.EntityHandler) *mux.Router {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Get the value of SECRET_KEY from the environment
	secretKey := os.Getenv("SECRET_KEY")

	r := mux.NewRouter()
	s := r.PathPrefix("/api").Subrouter()

	// Пользовательские маршруты
	r.HandleFunc("/users/register", func(w http.ResponseWriter, r *http.Request) {
		userHandler.RegisterHandler(w, r, "users")
	}).Methods("POST")

	r.HandleFunc("/restaurants/register", func(w http.ResponseWriter, r *http.Request) {
		restaurantHandler.RegisterHandler(w, r, "restaurants")
	}).Methods("POST")
	r.HandleFunc("/users/login", func(w http.ResponseWriter, r *http.Request) {
		authHandler.LoginHandler(w, r, "users")
	}).Methods("POST")

	r.HandleFunc("/restaurants/login", func(w http.ResponseWriter, r *http.Request) {
		authHandler.LoginHandler(w, r, "restaurants")

	}).Methods("POST")

	r.HandleFunc("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		userHandler.GetEntityById(w, r, "users")
	}).Methods("GET")

	r.HandleFunc("/restaurants/{id}", func(w http.ResponseWriter, r *http.Request) {
		restaurantHandler.GetEntityById(w, r, "restaurants")
	}).Methods("GET")

	// Secure rout

	s.Use(auth.AuthMiddleware([]byte(secretKey)))
	s.HandleFunc("/getall", userHandler.GetAllUsers).Methods("GET")
	s.HandleFunc("/users/me", func(w http.ResponseWriter, r *http.Request) {
		userHandler.GetEntity(w, r)
	}).Methods("GET")

	s.HandleFunc("/restaurants/me", func(w http.ResponseWriter, r *http.Request) {
		restaurantHandler.GetEntity(w, r)
	}).Methods("GET")

	s.HandleFunc("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		userHandler.UpdateEntity(w, r, "users")
	}).Methods("PUT")

	s.HandleFunc("/restaurants/{id}", func(w http.ResponseWriter, r *http.Request) {
		restaurantHandler.UpdateEntity(w, r, "restaurants")
	}).Methods("PUT")

	s.HandleFunc("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		userHandler.DeleteEntity(w, r, "users")
	}).Methods("DELETE")

	s.HandleFunc("/restaurants/{id}", func(w http.ResponseWriter, r *http.Request) {
		restaurantHandler.DeleteEntity(w, r, "restaurants")
	}).Methods("DELETE")

	return r
}
