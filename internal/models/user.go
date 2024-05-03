package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/go-playground/validator.v9"
)

type UserCredentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// User структура, представляющая пользователя
type User struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Email       string             `json:"email" bson:"email" validate:"required,email"`
	Password    string             `json:"password" bson:"password" validate:"required,min=6"`
	Surname     string             `json:"surname" bson:"surname" validate:"required"`
	Name        string             `json:"name" bson:"name" validate:"required"`
	Age         int                `json:"age" bson:"age" validate:"required,gte=0,lte=130"`
	Phone       string             `json:"phone" bson:"phone" validate:"required,len=11"`
	Interests   string             `json:"interests" bson:"interests" validate:"max=1000"`
	Description string             `json:"description" bson:"description" validate:"max=1000"`
	Avatar      string             `json:"avatar" bson:"avatar" validate:"max=1000"`
	Banned      bool               `json:"banned" bson:"banned,omitempty"`
	BanReason   string             `json:"ban_reason" bson:"banReason,omitempty"`
	Roles       string             `json:"roles" bson:"roles,omitempty"`
}

// Validate выполняет валидацию полей пользователя
func (u *User) Validate() error {
	validate := validator.New()
	return validate.Struct(u)
}