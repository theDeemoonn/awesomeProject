package auth

import (
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// JWTClaims структура для утверждений JWT
type JWTClaims struct {
	UserID primitive.ObjectID `json:"user_id"`
	Email  string             `json:"email"`
	Roles  string             `json:"roles"`
	jwt.StandardClaims
}

// JWTConfig конфигурация для JWT
type JWTConfig struct {
	SecretKey string        // Секретный ключ для подписи токена
	Duration  time.Duration // Продолжительность действия токена
}

// NewJWTConfig создает конфигурацию для JWT
func NewJWTConfig(secretKey string, duration time.Duration) JWTConfig {
	return JWTConfig{
		SecretKey: secretKey,
		Duration:  duration,
	}
}

// GenerateToken генерирует новый JWT токен для указанного пользователя
func GenerateToken(user_id primitive.ObjectID, email, roles, secretKey string, duration time.Duration) (string, error) {
	claims := JWTClaims{
		UserID: user_id,
		Email:  email,
		Roles:  roles,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(duration).Unix(),
			Issuer:    "your_application_name",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", errors.Wrap(err, "failed to sign the JWT token")
	}

	return signedToken, nil
}

// ValidateToken проверяет и декодирует JWT токен
func ValidateToken(signedToken, secretKey string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(signedToken, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, errors.Wrap(err, "failed to parse JWT token")
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, errors.New("invalid JWT token or claims")
	}
}
