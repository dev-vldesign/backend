package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Notification struct {
    ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    UserID     primitive.ObjectID `bson:"user_id" json:"user_id"`   
    SenderID   primitive.ObjectID `bson:"sender_id" json:"sender_id"` 
    Title      string             `bson:"title" json:"title"`
    Message    string             `bson:"message" json:"message"`
    IsSeen     bool               `bson:"is_seen" json:"is_seen"`
    CreatedAt  time.Time          `bson:"created_at" json:"created_at"`

    ContextID    *primitive.ObjectID `bson:"context_id,omitempty" json:"context_id,omitempty"`   
    ContextType  string              `bson:"context_type,omitempty" json:"context_type,omitempty"`
}