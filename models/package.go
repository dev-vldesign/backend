package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Package represents a bundle of features or a service level.
type Package struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name        string             `bson:"name" json:"name" validate:"required"` // e.g., "Basic Package", "Premium Package"
	Description string             `bson:"description" json:"description"`       // Description of what the package offers

	// Features map: The key is a unique identifier for the feature (e.g., "designerZoomCall"),
	// and the value is a descriptive text (e.g., "One Call", "Unlimited", "Included").
	// This makes features dynamic and textual.
	Features map[string]string `bson:"features" json:"features"`

	CreatedAt time.Time `bson:"created_at" json:"createdAt"`
	UpdatedAt time.Time `bson:"updated_at" json:"updatedAt"`
}
