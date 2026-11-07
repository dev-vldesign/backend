package models

import (
    "time"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

type Order struct {
    ID            primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
    ProjectID     primitive.ObjectID  `bson:"project_id" json:"project_id"`
    CustomerID    primitive.ObjectID  `bson:"customer_id" json:"customer_id"`
    TotalAmount   float64             `bson:"total_amount" json:"total_amount"`
    Status        string              `bson:"status" json:"status"`
    PaymentMethod string              `bson:"payment_method" json:"payment_method"`
    PaymentStatus string              `bson:"payment_status" json:"payment_status"`
    CreatedAt     time.Time           `bson:"created_at" json:"created_at"`
    UpdatedAt     time.Time           `bson:"updated_at" json:"updated_at"`
    Extras        []ExtraFeature      `bson:"extras" json:"extras"`
    CouponCode    *string             `bson:"coupon_code,omitempty" json:"coupon_code,omitempty"`
    Discount      float64             `bson:"discount,omitempty" json:"discount,omitempty"`
}

type ExtraFeature struct {
    Name  string  `bson:"name" json:"name"`
    Price float64 `bson:"price" json:"price"`
}
