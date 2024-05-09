package auth

import (
	"github.com/golang-jwt/jwt/v4"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"time"
)

// JWTClaims структура для утверждений JWT
type JWTClaims struct {
	UserID     primitive.ObjectID `json:"user_id"`
	Email      string             `json:"email"`
	Roles      string             `json:"roles"`
	EntityType string             `json:"entity_type"`
	jwt.RegisteredClaims
}
type Authenticatable interface {
	GetEmail() string
	GetPassword() string
	GetRoles() string
	GetID() primitive.ObjectID
	GetCollectionName() string

	GetCustomData() map[string]interface{}
}

// GenerateToken создает и возвращает доступный и рефреш JWT токен для пользователя.
func GenerateToken(entity Authenticatable, secretKey []byte, refreshTokenSecret []byte) (string, string, error) {

	accessClaims := &JWTClaims{
		UserID:     entity.GetID(),
		Email:      entity.GetEmail(),
		Roles:      entity.GetRoles(),
		EntityType: entity.GetCollectionName(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)), // 15 minutes
			Issuer:    "food&friends",
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	signedAccessToken, err := accessToken.SignedString(secretKey)

	if err != nil {
		return "", "", err
	}

	refreshClaims := &JWTClaims{
		UserID:     entity.GetID(),
		Email:      entity.GetEmail(),
		Roles:      entity.GetRoles(),
		EntityType: entity.GetCollectionName(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)), // 7 days
			Issuer:    "food&friends",
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	signedRefreshToken, err := refreshToken.SignedString(refreshTokenSecret)
	if err != nil {
		return "", "", err
	}

	return signedAccessToken, signedRefreshToken, nil
}

// ValidateToken проверяет и декодирует JWT токен
func ValidateToken(signedToken, secretKey string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(signedToken, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	if err != nil {
		log.Printf("Error parsing JWT token: %v", err)
		return nil, errors.Wrap(err, "failed to parse JWT token")
	}

	if claims, ok := token.Claims.(*JWTClaims); ok {
		if !token.Valid {
			log.Println("Token is not valid")
			return nil, errors.New("invalid JWT token")
		}
		return claims, nil
	} else {
		log.Println("Failed to cast claims to JWTClaims")
		return nil, errors.New("invalid JWT token or claims")
	}
}
