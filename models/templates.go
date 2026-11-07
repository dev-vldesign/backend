package models
// models/folder.go

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Folder struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	Name      string             `bson:"name" json:"name"`
	UserID    primitive.ObjectID `bson:"userId" json:"userId"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
}

type File struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	Name      string             `bson:"name" json:"name"`
	Size      int64              `bson:"size" json:"size"`
	Extension string             `bson:"extension" json:"extension"`
	FolderID  string             `bson:"folderId" json:"folderId"`
	Path      string             `bson:"path" json:"path"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
}