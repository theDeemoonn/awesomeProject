package main

import (
	"awesomeProject/internal/router"
	"awesomeProject/internal/services"
	"context"
	"crypto/rand"
	"fmt"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"net/http"
	"os"
	"time"

	"awesomeProject/internal/handlers"
	"awesomeProject/pkg/mongodb"
)

func GenerateRandomSecret(length int) ([]byte, error) {
	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}
	return randomBytes, nil
}

func main() {
	if err := godotenv.Load("/Users/dima/go/src/awesomeProject/.env"); err != nil {
		log.Fatal("Error loading .env file")
	}
	// Загрузка конфигурации
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		log.Fatal("MONGO_URI not defined in environment variables")
	}

	secretKey := os.Getenv("SECRET_KEY")
	if secretKey == "" {
		log.Fatal("JWT_SECRET_KEY is not set in environment variables")
	}

	refreshTokenSecret, err := GenerateRandomSecret(32) // Генерация 256-битного секрета для рефреш токенов
	if err != nil {
		log.Fatal("Failed to generate refresh token secret:", err)
	}

	// Подключение к MongoDB
	client, err := mongodb.NewMongoClient(mongodb.MongoDBConfig{
		URI:            mongoURI,
		ConnectTimeout: 10 * time.Second,
	})
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func(client *mongo.Client, ctx context.Context) {
		err := client.Disconnect(ctx)
		if err != nil {
			log.Fatalf("Failed to disconnect from MongoDB: %v", err)
		}
	}(client, context.Background())

	// Инициализация сервисов
	userService := services.NewUserService(client, "food", "users")

	// Инициализация обработчиков
	userHandler := handlers.NewUserHandler(userService)
	authHandler := handlers.NewAuthHandler(userService, []byte(secretKey), refreshTokenSecret)

	// Настройка роутинга
	r := router.InitializeRouter(userHandler, authHandler)
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
