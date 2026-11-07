package controllers

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Msaad668/vl-design-backend/db"
	"github.com/Msaad668/vl-design-backend/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreateFolder(c *gin.Context) {
	var input struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name is required"})
		return
	}

	userID, exists := c.Get("userID")

	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid userID format"})
		return
	}

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userObjectID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid userID"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	collection := db.GetCollection(db.Client, "folders")

	// Check if name already exists for that user
	filter := bson.M{"name": input.Name, "userId": userObjectID}
	var existing models.Folder
	err = collection.FindOne(context.Background(), filter).Decode(&existing)
	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Folder name already exists"})
		return
	}

	newFolder := models.Folder{
		ID:        primitive.NewObjectID(),
		Name:      input.Name,
		UserID:    userObjectID,
		CreatedAt: time.Now(),
	}

	_, err = collection.InsertOne(context.Background(), newFolder)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create folder"})
		return
	}

	c.JSON(http.StatusCreated, newFolder)
}

func UpdateFolderName(c *gin.Context) {
	folderID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(folderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid folder ID"})
		return
	}

	var input struct {
		NewName string `json:"newName" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "New name is required"})
		return
	}

	userId, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	collection := db.GetCollection(db.Client, "folders")

	// Ensure uniqueness of new name
	filter := bson.M{"name": input.NewName, "userId": userId}
	var existing models.Folder
	if err := collection.FindOne(context.Background(), filter).Decode(&existing); err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Folder name already exists"})
		return
	}

	update := bson.M{"$set": bson.M{"name": input.NewName}}
	_, err = collection.UpdateOne(context.Background(), bson.M{"_id": objID, "userId": userId}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update folder"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Folder name updated"})
}

func DeleteFolder(c *gin.Context) {
	folderID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(folderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid folder ID"})
		return
	}

	userId, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	collection := db.GetCollection(db.Client, "folders")

	res, err := collection.DeleteOne(context.Background(), bson.M{"_id": objID, "userId": userId})
	if err != nil || res.DeletedCount == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete folder"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Folder deleted"})
}

func GetTemplates(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page number"})
		return
	}

	const pageSize = 12
	skip := (page - 1) * pageSize

	collection := db.GetCollection(db.Client, "folders")

	opts := options.Find()
	opts.SetLimit(int64(pageSize))
	opts.SetSkip(int64(skip))
	opts.SetSort(bson.D{{"createdAt", -1}})

	cursor, err := collection.Find(context.Background(), bson.M{}, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch templates"})
		return
	}
	defer cursor.Close(context.Background())

	var templates []models.Folder
	if err := cursor.All(context.Background(), &templates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode templates"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"page":     page,
		"pageSize": pageSize,
		"items":    templates,
	})
}

func UploadFiles(c *gin.Context) {
	folderID := c.PostForm("folderId")
	if folderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "folderId is required"})
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid form data"})
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No files uploaded"})
		return
	}

	var savedFiles []models.File

	for _, file := range files {
		filename := filepath.Base(file.Filename)
		ext := filepath.Ext(filename)
		size := file.Size

		// Save the file to the uploads folder
		savePath := filepath.Join("uploads", filename)
		if err := c.SaveUploadedFile(file, savePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
			return
		}

		savedFiles = append(savedFiles, models.File{
			Name:      filename,
			Size:      size,
			Extension: ext,
			FolderID:  folderID,
			Path:      savePath,
			CreatedAt: time.Now(),
		})
	}

	// Insert files to MongoDB
	collection := db.GetCollection(db.Client, "files")
	var docs []interface{}
	for _, f := range savedFiles {
		docs = append(docs, f)
	}
	if _, err := collection.InsertMany(context.Background(), docs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Files uploaded successfully", "files": savedFiles})
}

func GetFolderDetails(c *gin.Context) {
	folderID := c.Param("id")
	if folderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Folder ID is required"})
		return
	}

	objID, err := primitive.ObjectIDFromHex(folderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid folder ID"})
		return
	}

	// Get folder
	var folder models.Folder
	folderCol := db.GetCollection(db.Client, "folders")
	err = folderCol.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&folder)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Folder not found"})
		return
	}

	// Get files in folder
	fileCol := db.GetCollection(db.Client, "files")
	cursor, err := fileCol.Find(context.Background(), bson.M{"folderId": folderID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch files"})
		return
	}
	defer cursor.Close(context.Background())

	var files []gin.H
	for cursor.Next(context.Background()) {
		var file models.File
		if err := cursor.Decode(&file); err == nil {
			files = append(files, gin.H{
				"name":      file.Name,
				"url":       strings.ReplaceAll(file.Path, "\\", "/"), // Use saved path
				"size":      file.Size,
				"extension": file.Extension,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"name":  folder.Name,
		"files": files,
	})
}

func UploadFilesToFolder(c *gin.Context) {
	folderID := c.Param("id")
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse form data"})
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no files provided"})
		return
	}

	// validate folder exists
	folderObjID, err := primitive.ObjectIDFromHex(folderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid folder ID"})
		return
	}

	folderCollection := db.GetCollection(db.Client, "folders")
	var folder models.Folder
	err = folderCollection.FindOne(context.Background(), bson.M{"_id": folderObjID}).Decode(&folder)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "folder not found"})
		return
	}

	var savedFiles []models.File
	uploadPath := "uploads"

	for _, file := range files {
		extension := filepath.Ext(file.Filename)
		newFileName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), file.Filename)
		dst := filepath.Join(uploadPath, newFileName)

		// Save file to disk
		if err := c.SaveUploadedFile(file, dst); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
			return
		}

		// Save metadata
		savedFiles = append(savedFiles, models.File{
			ID:        primitive.NewObjectID(),
			Name:      file.Filename,
			Size:      file.Size,
			Extension: extension,
			FolderID:  folderID,
			Path:      dst,
			CreatedAt: time.Now(),
		})
	}

	// Insert to DB
	fileCollection := db.GetCollection(db.Client, "files")
	var docs []interface{}
	for _, f := range savedFiles {
		docs = append(docs, f)
	}

	_, err = fileCollection.InsertMany(context.Background(), docs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file metadata"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "files uploaded", "files": savedFiles})
}
