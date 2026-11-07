package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Extra represents an additional feature, service, or product that can be added on.
type Extra struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name        string             `bson:"name" json:"name" validate:"required"` // e.g., "3D Renders", "Additional Revision"
	Description []string           `bson:"description" json:"description"`       // Bullet points describing the extra

	// PricesBySize defines the price of this extra for each applicable product size.
	// This makes extra pricing dynamic based on the size of the main product.
	PricesBySize []ExtraPriceBySize `bson:"prices_by_size" json:"pricesBySize" validate:"required,min=1"`

	// FeatureKey is crucial for filtering. It's a unique identifier (like a slug)
	// that represents the core functionality this extra provides.
	// This key should match a potential key in a Package's 'Features' map.
	FeatureKey string `bson:"feature_key" json:"featureKey" validate:"required"`

	CreatedAt time.Time `bson:"created_at" json:"createdAt"`
	UpdatedAt time.Time `bson:"updated_at" json:"updatedAt"`
}

// ExtraPriceBySize defines the price of an extra for a specific product size.
type ExtraPriceBySize struct {
	SizeName string  `bson:"size_name" json:"sizeName" validate:"required"` // Matches ProductSize.Name (e.g., "Small", "Medium")
	Price    float64 `bson:"price" json:"price" validate:"required,gte=0"`  // Price of this extra for this specific size
}
