package services

import (
	"awesomeProject/internal/auth"
	"awesomeProject/internal/models"
	"context"
	"crypto/rand"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
	"log"
	"os"
)

// Register регистрирует нового пользователя в системе
func (s *EntityService) Register(ctx context.Context, auth auth.Authenticatable) (string, error) {
	collectionName := auth.GetCollectionName() // Получение имени коллекции
	collection := s.db.Collection(collectionName)
	userData := auth.GetCustomData()
	//Проверка на уникальность email
	count, err := collection.CountDocuments(ctx, bson.M{"email": auth.GetEmail()})

	if err != nil {
		return "", errors.Wrap(err, "checking email failed")
	}
	if count > 0 {
		return "", errors.New("этот email уже зарегистрирован в системе")
	}

	// Хэширование пароля пользователя
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(auth.GetPassword()), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.Wrap(err, "hashing password failed")
	}
	userData["password"] = string(hashedPassword) // Добавляем хэшированный пароль

	// Добавление пользователя в базу данных

	result, err := collection.InsertOne(ctx, bson.M(userData))
	if err != nil {
		return "", errors.Wrap(err, "inserting user failed")
	}

	// Получение ID вставленного пользователя
	userID := result.InsertedID.(primitive.ObjectID).Hex()
	return userID, nil
}

// Authenticate проверяет учетные данные пользователя и возвращает токен, если успешно
func (s *EntityService) Authenticate(ctx context.Context, email, password string, auth auth.Authenticatable) (models.User, error) {
	collectionName := auth.GetCollectionName() // Получение имени коллекции
	collection := s.db.Collection(collectionName)
	var user models.User
	err := collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
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
func (s *EntityService) SaveRefreshToken(ctx context.Context, userID primitive.ObjectID, token string, auth auth.Authenticatable) error {
	collectionName := auth.GetCollectionName() // Получение имени коллекции
	collection := s.db.Collection(collectionName)
	_, err := collection.UpdateOne(ctx, bson.M{"_id": userID}, bson.M{"$set": bson.M{"refreshToken": token}})
	return err
}

// GenerateAndStoreToken генерирует токены и сохраняет рефреш токен в базе данных
func (s *EntityService) GenerateAndStoreToken(ctx context.Context, entity auth.Authenticatable, secretKey []byte, refreshTokenSecret []byte) (string, string, error) {
	collectionName := entity.GetCollectionName() // Получение имени коллекции
	collection := s.db.Collection(collectionName)
	accessToken, refreshToken, err := auth.GenerateToken(entity, secretKey, refreshTokenSecret)
	if err != nil {
		return "", "", err
	}

	update := bson.M{"$set": bson.M{"refreshToken": refreshToken}}
	_, err = collection.UpdateOne(ctx, bson.M{"_id": entity.GetID()}, update)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// GenerateRandomSecret генерирует случайный секрет заданной длины
func GenerateRandomSecret(length int) ([]byte, error) {
	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}
	return randomBytes, nil
}

// GetSecretKeys возвращает секретные ключи из переменных окружения
func (s *EntityService) GetSecretKeys() ([]byte, []byte) {
	if err := godotenv.Load(".env"); err != nil {
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

	return []byte(secretKey), refreshTokenSecret
}

func (s *EntityService) AuthenticateAndGenerateTokens(ctx context.Context, entity auth.Authenticatable, secretKey []byte, refreshTokenSecret []byte) (string, string, error) {
	// Directly authenticate using the provided entity and not converting it to a user model
	user, err := s.Authenticate(ctx, entity.GetEmail(), entity.GetPassword(), entity)
	if err != nil {
		return "", "", err
	}

	// Since the entity is authenticated, proceed to generate and store tokens
	accessToken, refreshToken, err := s.GenerateAndStoreToken(ctx, &user, secretKey, refreshTokenSecret)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}
