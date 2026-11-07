package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	FirstName string             `bson:"first_name" json:"first_name"`
	LastName  string             `bson:"last_name" json:"last_name"`
	Email     string             `bson:"email" json:"email"`
	Role      string             `bson:"role" json:"role"`                       // "admin", "customer", "designer"
	Phone     string             `bson:"phone,omitempty" json:"phone,omitempty"` // Optional
	Password  string             `bson:"password" json:"-"`
	Address   string             `bson:"address,omitempty" json:"address,omitempty"` // Optional
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	Image     string             `bson:"image" json:"image"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

type SignupInput struct {
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	Phone     string `json:"phone,omitempty"`
	Address   string `json:"address,omitempty"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UpdatePasswordInput struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}
