package models

import (
    "time"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

type Coupon struct {
    ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    Code        string             `bson:"code" json:"code"`
    Discount    float64            `bson:"discount" json:"discount"`
    ExpiryDate  *time.Time         `bson:"expiry_date,omitempty" json:"expiry_date,omitempty"`
    IsActive    bool               `bson:"is_active" json:"is_active"`
    CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
    UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}
