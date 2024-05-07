package router

import (
	"awesomeProject/internal/auth"
	"awesomeProject/internal/handlers"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"log"
	"os"
)

// InitializeRouter настраивает и возвращает роутер
func InitializeRouter(userHandler *handlers.UserHandler, authHandler *handlers.AuthHandler) *mux.Router {
	if err := godotenv.Load("/Users/dima/go/src/awesomeProject/.env"); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Get the value of SECRET_KEY from the environment
	secretKey := os.Getenv("SECRET_KEY")

	r := mux.NewRouter()
	s := r.PathPrefix("/api").Subrouter()

	// Пользовательские маршруты
	r.HandleFunc("/users/register", userHandler.RegisterUser).Methods("POST")
	r.HandleFunc("/users/login", authHandler.LoginHandler).Methods("POST")

	// Secure rout

	s.Use(auth.AuthMiddleware([]byte(secretKey)))
	s.HandleFunc("/getall", userHandler.GetAllUsers).Methods("GET")
	s.HandleFunc("/user", userHandler.GetUser).Methods("GET")
	s.HandleFunc("/users/{id}", userHandler.DeleteUser).Methods("DELETE")
	s.HandleFunc("/users/{id}", userHandler.UpdateUser).Methods("PUT")

	return r
}
