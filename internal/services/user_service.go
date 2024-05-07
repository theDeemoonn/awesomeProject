package services

import (
	"awesomeProject/internal/auth"
	"context"
	"crypto/rand"
	"github.com/joho/godotenv"

	"log"
	"os"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"

	"awesomeProject/internal/models"
	_ "awesomeProject/pkg/mongodb"
)

// UserService структура сервиса пользователей
type UserService struct {
	db *mongo.Collection
}

// SecretKey - ключ для подписи JWT токена
var SecretKey []byte

func init() {
	// Load environment variables from .env file
	if err := godotenv.Load("/Users/dima/go/src/awesomeProject/.env"); err != nil {
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

// NewUserService создает новый экземпляр UserService
func NewUserService(client *mongo.Client, dbName, collName string) *UserService {
	return &UserService{
		db: client.Database(dbName).Collection(collName),
	}
}

// GetUser возвращает пользователя по его ID
func (s *UserService) GetUser(ctx context.Context, userID string) (*models.User, error) {
	var user models.User
	id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err // не удалось преобразовать userID в ObjectID
	}

	filter := bson.M{"_id": id}
	if err := s.db.FindOne(ctx, filter).Decode(&user); err != nil {
		return nil, err // пользователь не найден или другая ошибка запроса
	}

	return &user, nil
}

// Register регистрирует нового пользователя в системе
func (s *UserService) Register(ctx context.Context, user *models.User) (string, error) {
	//Проверка на уникальность email
	count, err := s.db.CountDocuments(ctx, bson.M{"email": user.Email})
	if err != nil {
		return "", errors.Wrap(err, "finding user failed")
	}
	if err != nil {
		return "", errors.Wrap(err, "checking email failed")
	}
	if count > 0 {
		return "", errors.New("этот email уже зарегистрирован в системе")
	}

	// Хэширование пароля пользователя
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.Wrap(err, "hashing password failed")
	}
	user.Password = string(hashedPassword)

	// Добавление пользователя в базу данных
	result, err := s.db.InsertOne(ctx, user)
	if err != nil {
		return "", errors.Wrap(err, "inserting user failed")
	}

	// Получение ID вставленного пользователя
	userID := result.InsertedID.(primitive.ObjectID).Hex()
	return userID, nil
}

// Authenticate проверяет учетные данные пользователя и возвращает токен, если успешно

func (s *UserService) Authenticate(ctx context.Context, email, password string) (models.User, error) {
	var user models.User
	err := s.db.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		return models.User{}, err // Пользователь не найден или другая ошибка запроса
	}

	// Проверка пароля
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return models.User{}, errors.New("invalid password") // Неверный пароль
	}

	return user, nil
}

// SaveRefreshToken сохраняет рефреш токен в документе пользователя
func (s *UserService) SaveRefreshToken(ctx context.Context, userID primitive.ObjectID, token string) error {
	_, err := s.db.UpdateOne(ctx, bson.M{"_id": userID}, bson.M{"$set": bson.M{"refreshToken": token}})
	return err
}

// UpdateUser обновляет данные пользователя
func (s *UserService) UpdateUser(ctx context.Context, userID string, user *models.User) error {
	// Проверка на уникальность email
	count, err := s.db.CountDocuments(ctx, bson.M{"email": user.Email, "_id": bson.M{"$ne": userID}})
	if err != nil {
		return errors.Wrap(err, "checking email failed")
	}
	if count > 0 {
		return errors.New("этот email уже зарегистрирован в системе")
	}

	// Преобразование userID в ObjectID
	id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err // не удалось преобразовать userID в ObjectID
	}

	// Обновление данных пользователя
	filter := bson.M{"_id": id}
	update := bson.M{"$set": user}
	_, err = s.db.UpdateOne(ctx, filter, update)
	if err != nil {
		return errors.Wrap(err, "updating user failed")
	}

	return nil
}

// DeleteUser удаляет пользователя из системы
func (s *UserService) DeleteUser(ctx context.Context, userID string) error {
	// Преобразование userID в ObjectID
	id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err // не удалось преобразовать userID в ObjectID
	}

	// Удаление пользователя
	filter := bson.M{"_id": id}
	_, err = s.db.DeleteOne(ctx, filter)
	if err != nil {
		return errors.Wrap(err, "deleting user failed")
	}

	return nil
}

// GetAllUsers возвращает всех пользователей
func (s *UserService) GetAllUsers(ctx context.Context) ([]models.User, error) {
	var users []models.User
	cursor, err := s.db.Find(ctx, bson.M{})
	if err != nil {
		return nil, errors.Wrap(err, "finding users failed")
	}
	if err := cursor.All(ctx, &users); err != nil {
		return nil, errors.Wrap(err, "decoding users failed")
	}
	return users, nil
}

// GenerateAndStoreToken генерирует токены и сохраняет рефреш токен в базе данных
func (s *UserService) GenerateAndStoreToken(ctx context.Context, user models.User, secretKey []byte, refreshTokenSecret []byte) (string, string, error) {
	// Вызов функции генерации токенов из пакета auth
	accessToken, refreshToken, err := auth.GenerateToken(user, secretKey, refreshTokenSecret)
	if err != nil {
		return "", "", err
	}

	// Обновление пользователя с новым рефреш токеном
	update := bson.M{"$set": bson.M{"refreshToken": refreshToken}}
	_, err = s.db.UpdateOne(ctx, bson.M{"_id": user.ID}, update)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func GenerateRandomSecret(length int) ([]byte, error) {
	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}
	return randomBytes, nil
}

func (s *UserService) GetSecretKeys() ([]byte, []byte) {
	if err := godotenv.Load("/Users/dima/go/src/awesomeProject/.env"); err != nil {
		log.Fatal("Error loading .env file")
	}

	secretKey := os.Getenv("SECRET_KEY")
	if secretKey == "" {
		log.Fatal("JWT_SECRET_KEY is not set in environment variables")
	}

	refreshTokenSecret, err := GenerateRandomSecret(32) // Генерация 256-битного секрета для рефреш токенов
	if err != nil {
		log.Fatal("Failed to generate refresh token secret:", err)
	}

	if secretKey == "" || string(refreshTokenSecret) == "" {
		log.Fatal("Secret keys must be set as environment variables")
	}

	return []byte(secretKey), refreshTokenSecret
}
