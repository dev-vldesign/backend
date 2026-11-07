package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Pricing defines the specific price for a given Product-Package-Size combination.
type Pricing struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	ProductID primitive.ObjectID `bson:"product_id" json:"productId" validate:"required"` // Reference to the Product
	PackageID primitive.ObjectID `bson:"package_id" json:"packageId" validate:"required"` // Reference to the Package
	SizeName  string             `bson:"size_name" json:"sizeName" validate:"required"`   // Matches ProductSize.Name
	Price     float64            `bson:"price" json:"price" validate:"required,gte=0"`    // The price for this specific combo
	CreatedAt time.Time          `bson:"created_at" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updatedAt"`
}
