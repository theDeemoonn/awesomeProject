package main

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"time"

	"awesomeProject/internal/handlers"
	"awesomeProject/internal/services"
	"awesomeProject/pkg/mongodb"
	"github.com/gorilla/mux"
)

func main() {
	if err := godotenv.Load("/Users/dima/go/src/awesomeProject/.env"); err != nil {
		log.Fatal("Error loading .env file")
	}
	// Загрузка конфигурации
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		log.Fatal("MONGO_URI not defined in environment variables")
	}

	// Подключение к MongoDB
	client, err := mongodb.NewMongoClient(mongodb.MongoDBConfig{
		URI:            mongoURI,
		ConnectTimeout: 10 * time.Second,
	})
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(context.Background())

	// Инициализация сервисов
	userService := services.NewUserService(client, "food", "users")

	// Инициализация обработчиков
	userHandler := handlers.NewUserHandler(userService)
	authHandler := handlers.NewAuthHandler(userService)

	// Настройка роутинга
	r := mux.NewRouter()
	r.HandleFunc("/users/register", userHandler.RegisterUser).Methods("POST")
	r.HandleFunc("/users/login", authHandler.LoginHandler).Methods("POST")
	r.HandleFunc("/users/{id}", userHandler.GetUser).Methods("GET")

	// Настройка и запуск HTTP сервера
	httpServer := &http.Server{
		Handler:      r,
		Addr:         ":8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	fmt.Println("Starting server on port 8080")
	log.Fatal(httpServer.ListenAndServe())
}
