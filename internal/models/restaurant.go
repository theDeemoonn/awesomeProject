package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/go-playground/validator.v9"
)

// Restaurant структура, представляющая ресторан
type Restaurant struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	Name         string             `bson:"name" validate:"required"`
	AveragePrice int                `bson:"averagePrice" validate:"required,gt=0"`
	Description  string             `bson:"description" validate:"required"`
	Category     string             `bson:"category" validate:"required"`
	OGRN         string             `bson:"ogrn" validate:"required,len=13"` // Проверка длины
	INN          string             `bson:"inn" validate:"required,len=10"`  // Проверка длины
	Address      string             `bson:"address" validate:"required"`
	Avatar       string             `bson:"avatar,omitempty"`
	Email        string             `bson:"email" validate:"required,email"`
	Password     string             `bson:"password" validate:"required,min=6"`
	Phone        string             `bson:"phone" validate:"required,len=11"`
	Banned       bool               `bson:"banned,omitempty"`
	BanReason    string             `bson:"banReason,omitempty"`
}

// ValidateRestaurant проводит валидацию полей ресторана
func ValidateRestaurant(restaurant *Restaurant) error {
	validate := validator.New()
	// Custom validators can be added here for OGRN and INN if needed
	return validate.Struct(restaurant)
}
