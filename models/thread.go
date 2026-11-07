package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Thread struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ProjectID primitive.ObjectID `bson:"project_id" json:"project_id"`
	AuthorID  primitive.ObjectID `bson:"author_id" json:"author_id"`
	Role      string             `bson:"role" json:"role"`
	Message   string             `bson:"message" json:"message"`
	Documents []Document         `bson:"documents,omitempty" json:"documents,omitempty"`
	Replies   []Reply            `bson:"replies,omitempty" json:"replies,omitempty"`
	IsDraft   bool               `bson:"is_draft" json:"is_draft"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

type Reply struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ThreadID   primitive.ObjectID `bson:"thread_id" json:"thread_id"`
	AuthorID   primitive.ObjectID `bson:"author_id" json:"author_id"`
	AuthorName string             `bson:"author_name" json:"author_name"`
	Role       string             `bson:"role" json:"role"`
	Message    string             `bson:"message" json:"message"`
	Documents  []Document         `bson:"documents,omitempty" json:"documents,omitempty"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time          `bson:"updated_at" json:"updated_at"`
}

type Document struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name       string             `bson:"name" json:"name"`
	Size       string             `bson:"size" json:"size"`
	URL        string             `bson:"url" json:"url"`
	UploadedBy primitive.ObjectID `bson:"uploaded_by" json:"uploaded_by"`
}
