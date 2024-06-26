package services

import (
	"awesomeProject/internal/auth"
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"reflect"

	"log"
	"os"

	"awesomeProject/internal/models"
	_ "awesomeProject/pkg/mongodb"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// EntityService структура сервиса пользователей
type EntityService struct {
	db             *mongo.Database
	entityCollName string
}

const (
	EntityTypeUser       = "users"
	EntityTypeRestaurant = "restaurants"
)

// SecretKey - ключ для подписи JWT токена
var SecretKey []byte

func init() {
	// Load environment variables from .env file
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Get the value of SECRET_KEY from the environment
	secretKey := os.Getenv("SECRET_KEY")
	if secretKey == "" {
		log.Fatal("SECRET_KEY not defined in environment variables")
	}

	// Convert the secretKey string to []byte
	SecretKey = []byte(secretKey)
}

// NewEntityService создает новый экземпляр EntityService
func NewEntityService(client *mongo.Client, dbName, entityCollName string) *EntityService {
	return &EntityService{
		db:             client.Database(dbName),
		entityCollName: entityCollName,
	}
}

// GetEntity возвращает сущность по ее ID и типу
func (s *EntityService) GetEntity(ctx context.Context, entityID string, entityType string) (interface{}, error) {
	var result interface{}

	id, err := primitive.ObjectIDFromHex(entityID)
	if err != nil {
		return nil, err // не удалось преобразовать entityID в ObjectID
	}

	filter := bson.M{"_id": id}
	switch entityType {
	case EntityTypeUser:
		result = &models.User{}
		if err := s.db.Collection(s.entityCollName).FindOne(ctx, filter).Decode(result); err != nil {
			return nil, err // сущность не найдена или другая ошибка запроса
		}
	case EntityTypeRestaurant:
		result = &models.Restaurant{}
		if err := s.db.Collection(s.entityCollName).FindOne(ctx, filter).Decode(result); err != nil {
			return nil, err // сущность не найдена или другая ошибка запроса
		}
	default:
		return nil, fmt.Errorf("unknown entity type: %s", entityType)
	}

	return result, nil
}

// UpdateEntity обновляет данные сущности по ее ID и типу
func (s *EntityService) UpdateEntity(ctx context.Context, entityID string, entity interface{}, entityType string) error {
	var collectionName string

	// Определение коллекции на основе типа сущности
	switch entityType {
	case "users":
		collectionName = "users"
	case "restaurant":
		collectionName = "restaurants"
	default:
		return fmt.Errorf("unknown entity type: %s", entityType)
	}

	// Проверка на уникальность email (предполагается, что entity имеет поле Email)
	email := reflect.ValueOf(entity).Elem().FieldByName("Email").String()
	count, err := s.db.Collection(collectionName).CountDocuments(ctx, bson.M{"email": email, "_id": bson.M{"$ne": entityID}})
	if err != nil {
		return errors.Wrap(err, "checking email failed")
	}
	if count > 0 {
		return errors.New("this email is already registered in the system")
	}

	// Преобразование entityID в ObjectID
	id, err := primitive.ObjectIDFromHex(entityID)
	if err != nil {
		return err // не удалось преобразовать entityID в ObjectID
	}

	// Обновление данных сущности
	filter := bson.M{"_id": id}
	update := bson.M{"$set": entity}
	_, err = s.db.Collection(collectionName).UpdateOne(ctx, filter, update)
	if err != nil {
		return errors.Wrap(err, "updating entity failed")
	}

	return nil
}

// ChangePassword изменяет пароль сущности по ее ID и типу
func (s *EntityService) ChangePassword(ctx context.Context, entityID string, oldPassword, newPassword string, entity auth.Authenticatable) error {
	collectionName := entity.GetCollectionName()
	collection := s.db.Collection(collectionName)

	// Преобразование entityID в ObjectID
	objID, err := primitive.ObjectIDFromHex(entityID)
	if err != nil {
		return errors.Wrap(err, "invalid entity ID")
	}

	// Получение сущности из базы данных
	var storedEntity auth.Authenticatable
	err = collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&storedEntity)
	if err != nil {
		return errors.Wrap(err, "entity not found")
	}

	// Проверка старого пароля
	if err := bcrypt.CompareHashAndPassword([]byte(storedEntity.GetPassword()), []byte(oldPassword)); err != nil {
		return errors.New("invalid old password")
	}

	// Хэширование нового пароля
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.Wrap(err, "hashing new password failed")
	}

	// Обновление пароля в базе данных
	update := bson.M{"$set": bson.M{"password": string(hashedPassword)}}
	_, err = collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return errors.Wrap(err, "updating password failed")
	}

	return nil
}

// DeleteEntity удаляет сущность по ее ID и типу
func (s *EntityService) DeleteEntity(ctx context.Context, entityID string, entityType string) error {
	var collectionName string

	// Определение коллекции на основе типа сущности
	switch entityType {
	case "user":
		collectionName = "users" // Имя коллекции для пользователей
	case "restaurant":
		collectionName = "restaurants" // Имя коллекции для ресторанов
	default:
		return fmt.Errorf("unknown entity type: %s", entityType)
	}

	// Преобразование entityID в ObjectID
	id, err := primitive.ObjectIDFromHex(entityID)
	if err != nil {
		return err // не удалось преобразовать entityID в ObjectID
	}

	// Удаление сущности
	filter := bson.M{"_id": id}
	_, err = s.db.Collection(collectionName).DeleteOne(ctx, filter)
	if err != nil {
		return errors.Wrap(err, "deleting entity failed")
	}

	return nil
}

// GetAllUsers возвращает всех пользователей
func (s *EntityService) GetAllUsers(ctx context.Context) ([]models.User, error) {
	var collectionName string
	var users []models.User
	cursor, err := s.db.Collection(collectionName).Find(ctx, bson.M{})
	if err != nil {
		return nil, errors.Wrap(err, "finding users failed")
	}
	if err := cursor.All(ctx, &users); err != nil {
		return nil, errors.Wrap(err, "decoding users failed")
	}
	return users, nil
}

// GetAllRestaurants возвращает всех ресторанов
func (s *EntityService) GetAllRestaurants(ctx context.Context) ([]models.Restaurant, error) {
	var collectionName string
	var restaurants []models.Restaurant
	cursor, err := s.db.Collection(collectionName).Find(ctx, bson.M{})
	if err != nil {
		return nil, errors.Wrap(err, "finding restaurants failed")
	}
	if err := cursor.All(ctx, &restaurants); err != nil {
		return nil, errors.Wrap(err, "decoding restaurants failed")
	}
	return restaurants, nil
}

// AddFavoriteRestaurant добавляет ресторан в список избранных у пользователя
func (s *EntityService) AddFavoriteRestaurant(ctx context.Context, userID primitive.ObjectID, restaurantID primitive.ObjectID) error {
	collection := s.db.Collection("users")

	// Обновление списка избранных ресторанов
	filter := bson.M{"_id": userID}
	update := bson.M{"$addToSet": bson.M{"favorites": restaurantID}}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return errors.Wrap(err, "adding favorite restaurant failed")
	}

	return nil
}

// GetFavoriteRestaurants возвращает список избранных ресторанов пользователя
func (s *EntityService) GetFavoriteRestaurants(ctx context.Context, userID primitive.ObjectID) ([]models.Restaurant, error) {
	userCollection := s.db.Collection("users")
	var user models.User

	// Получение данных пользователя
	err := userCollection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		return nil, errors.Wrap(err, "user not found")
	}

	restaurantCollection := s.db.Collection("restaurants")
	var favoriteRestaurants []models.Restaurant

	// Получение данных избранных ресторанов
	cursor, err := restaurantCollection.Find(ctx, bson.M{"_id": bson.M{"$in": user.Favorites}})
	if err != nil {
		return nil, errors.Wrap(err, "finding favorite restaurants failed")
	}
	if err := cursor.All(ctx, &favoriteRestaurants); err != nil {
		return nil, errors.Wrap(err, "decoding favorite restaurants failed")
	}

	return favoriteRestaurants, nil
}
