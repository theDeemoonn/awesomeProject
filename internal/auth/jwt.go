package auth

import (
	"awesomeProject/internal/models"
	"github.com/golang-jwt/jwt/v4"
	"time"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// JWTClaims структура для утверждений JWT
type JWTClaims struct {
	UserID primitive.ObjectID `json:"user_id"`
	Email  string             `json:"email"`
	Roles  string             `json:"roles"`
	jwt.RegisteredClaims
}

// GenerateToken генерирует новый JWT токен для указанного пользователя
func GenerateToken(user models.User, secretKey []byte) (string, error) {
	claims := JWTClaims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			Issuer:    "food&friends",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(secretKey)
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
