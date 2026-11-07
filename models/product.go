package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Product represents a core design product or service offered.
type Product struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name        string             `bson:"name" json:"name" validate:"required"`         // e.g., "Landscape Design"
	Description []string           `bson:"description" json:"description"`               // Bullet points describing the product
	Sizes       []ProductSize      `bson:"sizes" json:"sizes" validate:"required,min=1"` // Available sizes/tiers for this product
	CreatedAt   time.Time          `bson:"created_at" json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updatedAt"`
}

// ProductSize defines a specific size option for a Product.
type ProductSize struct {
	Name string  `bson:"name" json:"name" validate:"required"` // e.g., "Small", "Medium", "Large"
	Min  float64 `bson:"min" json:"min"`                       // Minimum dimension (e.g., square feet, meters)
	Max  float64 `bson:"max" json:"max"`                       // Maximum dimension
	Unit string  `bson:"unit" json:"unit"`                     // Unit of measurement (e.g., "sq ft", "sq m")
}
