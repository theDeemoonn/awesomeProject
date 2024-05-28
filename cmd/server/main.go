package main

import (
	"awesomeProject/internal/router"
	"awesomeProject/internal/services"
	"context"
	"crypto/rand"
	"fmt"
	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"net/http"
	"os"
	"time"

	"awesomeProject/internal/handlers"
	"awesomeProject/pkg/mongodb"
)

// @title Swagger Example API
// @version 1.0
// @description This is a sample server celler server.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api

func GenerateRandomSecret(length int) ([]byte, error) {
	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}
	return randomBytes, nil
}

// @Summary Show an account
// @Description get string by ID
// @ID get-string-by-int
// @Accept  json
// @Produce  json
// @Param id path int true "Account ID"
// @Success 200 {object} handlers.EntityHandler
// @Failure 400 {object} http.Error
// @Failure 404 {object} http.Error
// @Failure 500 {object} http.Error
// @Router /accounts/{id} [get]
func main() {
	if err := godotenv.Load(".env"); err != nil {
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

	// Имя коллекции для пользователей
	usersCollectionName := "users"
	restaurantsCollectionName := "restaurants"

	// Инициализация сервисов
	userService := services.NewEntityService(client, "food", usersCollectionName)
	restaurantService := services.NewEntityService(client, "food", restaurantsCollectionName)
	redisService := services.NewRedisService()

	// Инициализация обработчиков
	userHandler := handlers.NewEntityHandler(userService, redisService)
	restaurantHandler := handlers.NewEntityHandler(restaurantService, redisService)
	authHandler := handlers.NewAuthHandler(userService, []byte(secretKey), refreshTokenSecret)

	// Настройка роутинга
	r := router.InitializeRouter(userHandler, authHandler, restaurantHandler)
	// Добавление маршрута для документации Swagger
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)
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
