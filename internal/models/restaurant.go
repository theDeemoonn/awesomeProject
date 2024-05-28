package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/go-playground/validator.v9"
	"time"
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
	Hours        string             `json:"hours" bson:"hours"`
	Banned       bool               `bson:"banned,omitempty"`
	BanReason    string             `bson:"banReason,omitempty"`
	Roles        string             `json:"roles" bson:"roles,omitempty"`
	RefreshToken string             `json:"-"`
	Menu         []MenuItem         `json:"menu" bson:"menu"`
	Orders       []Order            `json:"orders" bson:"orders"`
	Reviews      []Review           `json:"reviews" bson:"reviews"`
}

// MenuItem представляет информацию о блюде в меню ресторана.
type MenuItem struct {
	ID           string  `json:"id" bson:"_id,omitempty"`
	RestaurantID string  `json:"restaurant_id" bson:"restaurant_id"`
	Name         string  `json:"name" bson:"name"`
	Description  string  `json:"description" bson:"description"`
	Price        float64 `json:"price" bson:"price"`
	Category     string  `json:"category" bson:"category"`
	ImageURL     string  `json:"image_url" bson:"image_url"`
}

// Order представляет информацию о заказе.
type Order struct {
	ID            string        `json:"id" bson:"_id,omitempty"`
	UserID        string        `json:"user_id" bson:"user_id"`
	RestaurantID  string        `json:"restaurant_id" bson:"restaurant_id"`
	Items         []OrderItem   `json:"items" bson:"items"`
	PaymentMethod PaymentMethod `json:"payment_method" bson:"payment_method"`
	Status        string        `json:"status" bson:"status"` // pending, confirmed, preparing, delivered
	CreatedAt     time.Time     `json:"created_at" bson:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at" bson:"updated_at"`
}

// OrderItem представляет информацию о позиции заказа.
type OrderItem struct {
	MenuItemID string `json:"menu_item_id" bson:"menu_item_id"`
	Quantity   int    `json:"quantity" bson:"quantity"`
}

// Review представляет отзыв о ресторане.
type Review struct {
	ID           string    `json:"id" bson:"_id,omitempty"`
	RestaurantID string    `json:"restaurant_id" bson:"restaurant_id"`
	UserID       string    `json:"user_id" bson:"user_id"`
	Rating       int       `json:"rating" bson:"rating"`
	Comment      string    `json:"comment" bson:"comment"`
	CreatedAt    time.Time `json:"created_at" bson:"created_at"`
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
