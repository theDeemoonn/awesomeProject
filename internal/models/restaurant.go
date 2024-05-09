package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/go-playground/validator.v9"
)

type RestaurantCredentials struct {
	Email        string `json:"email"`
	Password     string `json:"password"`
	RefreshToken string `bson:"refreshToken,omitempty"`
}

// Restaurant структура, представляющая ресторан
type Restaurant struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	Email        string             `json:"email" bson:"email" validate:"required,email"`
	Password     string             `json:"password" bson:"password" validate:"required,min=6"`
	Name         string             `bson:"name" validate:"required"`
	AveragePrice int                `bson:"averagePrice" validate:"required,gt=0"`
	Description  string             `bson:"description" validate:"required"`
	Category     string             `bson:"category" validate:"required"`
	OGRN         string             `bson:"ogrn" validate:"required,len=13"` // Проверка длины
	INN          string             `bson:"inn" validate:"required,len=10"`  // Проверка длины
	Address      string             `bson:"address" validate:"required"`
	Avatar       string             `bson:"avatar,omitempty"`
	Phone        string             `bson:"phone" validate:"required,len=11"`
	Banned       bool               `bson:"banned,omitempty"`
	BanReason    string             `bson:"banReason,omitempty"`
	Roles        string             `json:"roles" bson:"roles,omitempty"`
	RefreshToken string             `json:"-"`
}

// ValidateRestaurant проводит валидацию полей ресторана
func ValidateRestaurant(restaurant *Restaurant) error {
	validate := validator.New()
	// Custom validators can be added here for OGRN and INN if needed
	return validate.Struct(restaurant)
}

func (r *Restaurant) GetEmail() string          { return r.Email }
func (r *Restaurant) GetPassword() string       { return r.Password }
func (r *Restaurant) GetID() primitive.ObjectID { return r.ID }
func (r *Restaurant) GetRoles() string          { return r.Roles }
func (r *Restaurant) GetCollectionName() string {
	return "restaurants"
}

func (r *Restaurant) GetCustomData() map[string]interface{} {
	return map[string]interface{}{
		"email":        r.Email,
		"password":     r.Password,
		"name":         r.Name,
		"averagePrice": r.AveragePrice,
		"description":  r.Description,
		"category":     r.Category,
		"ogrn":         r.OGRN,
		"inn":          r.INN,
		"address":      r.Address,
		"avatar":       r.Avatar,
		"phone":        r.Phone,
		"banned":       r.Banned,
		"banReason":    r.BanReason,
		"roles":        r.GetRoles(), // Добавлено получение ролей
	}
}
