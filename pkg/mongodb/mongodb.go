package mongodb

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// MongoDBConfig структура для хранения конфигурации подключения к MongoDB
type MongoDBConfig struct {
	URI            string        // URI подключения к MongoDB
	ConnectTimeout time.Duration // Таймаут подключения
}

// NewMongoClient создаёт новый экземпляр клиента MongoDB
func NewMongoClient(cfg MongoDBConfig) (*mongo.Client, error) {
	// Настройка опций подключения

	clientOptions := options.Client().ApplyURI(cfg.URI).SetConnectTimeout(cfg.ConnectTimeout)

	// Создание клиента MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return nil, err
	}

	// Проверка подключения
	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		return nil, err
	}

	log.Println("Successfully connected and pinged MongoDB.")
	return client, nil
}

// DisconnectMongoClient отключает клиента от MongoDB
func DisconnectMongoClient(client *mongo.Client) error {
	if err := client.Disconnect(context.TODO()); err != nil {
		log.Printf("Failed to disconnect from MongoDB: %v", err)
		return err
	}
	log.Println("Disconnected from MongoDB.")
	return nil
}
