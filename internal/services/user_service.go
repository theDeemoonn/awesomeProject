package services

import (
	"context"
	"github.com/dgrijalva/jwt-go"
	"time"
	_ "time"

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
var SecretKey = []byte("secret")

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
func (s *UserService) Authenticate(ctx context.Context, email, password string) (string, error) {
	// Поиск пользователя по email
	var user models.User
	err := s.db.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		return "", errors.Wrap(err, "finding user failed")
	}

	// Проверка пароля
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", errors.Wrap(err, "invalid password")
	}

	// Генерация JWT токена (псевдокод, предполагает наличие функции GenerateJWT)
	token, err := GenerateJWT(user)
	if err != nil {
		return "", errors.Wrap(err, "generating JWT failed")
	}

	return token, nil
}

// GenerateJWT генерирует JWT токен для пользователя
// Это псевдокод и требует реализации в соответствии с вашей JWT библиотекой
func GenerateJWT(user models.User) (string, error) {
	// Создание нового токена
	token := jwt.New(jwt.SigningMethodHS256)

	// Установка пользовательских данных в токен
	claims := token.Claims.(jwt.MapClaims)
	claims["id"] = user.ID
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix() // Токен действителен в течение 24 часов

	// Подписание токена с использованием секретного ключа
	tokenString, err := token.SignedString(SecretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
