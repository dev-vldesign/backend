package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Portfolio struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID         primitive.ObjectID `bson:"user_id" json:"user_id"`
	Name           string             `bson:"name" json:"name"`
	ProfilePicture string             `bson:"profile_picture" json:"profile_picture"`
	Cover          string             `bson:"cover" json:"cover"`
	About          string             `bson:"about" json:"about"`
	Experience     string             `bson:"experience" json:"experience"`
	Nationality    string             `bson:"nationality" json:"nationality"`
	Tools          []string           `bson:"tools" json:"tools"` // ["Blender", "Autodesk Maya"]
	Projects       []PortfolioProject `bson:"projects" json:"projects"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at" json:"updated_at"`
}

type Image struct {
	ID   primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Path string             `bson:"path,omitempty" json:"path"`
}

type PortfolioProject struct {
	ID     primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title  string             `bson:"title" json:"title"`
	Images []Image           `bson:"images" json:"images"` // URLs to images
}
