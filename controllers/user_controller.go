package controllers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/Msaad668/vl-design-backend/db"
	"github.com/Msaad668/vl-design-backend/models"
	"github.com/Msaad668/vl-design-backend/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

func GetUserInfo(c *gin.Context) {
	email, exists := c.Get("email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userCollection := db.GetCollection(db.Client, "users")

	var user models.User
	err := userCollection.FindOne(context.TODO(), bson.M{"email": email}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         user.ID,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"email":      user.Email,
		"phone":      user.Phone,
		"address":    user.Address,
		"role":       user.Role,
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
		"image":      user.Image,
	})
}

func UpdateUserInfo(c *gin.Context) {
	userIDVal, exists := c.Get("userID")

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// validate userID
	stringUserID, ok := userIDVal.(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid userID format"})
		return
	}

	userObjectID, err := primitive.ObjectIDFromHex(stringUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid userID"})
		return
	}

	// check if user exists
	usersCollection := db.GetCollection(db.Client, "users")
	var user models.User

	if err := usersCollection.FindOne(context.Background(), bson.M{"_id": userObjectID}).Decode(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "This user doesn't exist"})
	}

	updateData := bson.M{}

	if c.PostForm("role") != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You are not allowed to change your role"})
		return
	}

	if first_name := c.PostForm("first_name"); first_name != "" {
		updateData["first_name"] = c.PostForm("first_name")
	}

	if last_name := c.PostForm("last_name"); last_name != "" {
		updateData["last_name"] = c.PostForm("last_name")
	}

	if email := c.PostForm("email"); email != "" {
		updateData["email"] = c.PostForm("email")
	}

	if phone := c.PostForm("phone"); phone != "" {
		updateData["phone"] = c.PostForm("phone")
	}

	// handle the image if it exists
	file, handler, err := c.Request.FormFile("image")
	if err == nil {
		defer file.Close()

		allowedTypes := map[string]bool{
			"image/jpeg": true,
			"image/png":  true,
			"image/gif":  true,
			"image/webp": true,
			// Add other image types you want to support
		}

		contentType := handler.Header.Get("Content-Type")
		if !allowedTypes[contentType] {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":         "Only images are allowed (JPEG, PNG, GIF, WEBP)",
				"received_type": contentType,
			})
			return
		}

		// handle file extension
		ext := filepath.Ext(handler.Filename)
		newFilename := fmt.Sprintf("%s%s", stringUserID, ext)

		// Create the file path (OS-specific)
		filePath := filepath.Join("uploads", "users", newFilename)

		// Ensure the uploads/users directory exists
		err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to create directory",
				"message": err.Error(),
			})
			return
		}

		// Create the file
		out, err := os.Create(filePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to save image",
				"message": err.Error(),
			})
			return
		}
		defer out.Close()

		// Copy the uploaded file content into the created file
		_, err = io.Copy(out, file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write image content"})
			return
		}

		// Normalize the path for database storage (convert to forward slashes)
		normalizedPath := "/" + filepath.ToSlash(filePath) // Adds leading slash and converts to /

		// Store normalized path in update data
		updateData["image"] = normalizedPath
	} else if err != http.ErrMissingFile {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error processing image upload"})
		return
	} else {

	}

	updateData["updated_at"] = time.Now()

	var updatedUser models.User
	err = usersCollection.FindOneAndUpdate(
		context.Background(),
		bson.M{"_id": userObjectID},
		bson.M{"$set": updateData},
		options.FindOneAndUpdate().SetReturnDocument(options.After), // Return the updated document
	).Decode(&updatedUser)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User updated successfully",
		"user":    updatedUser,
	})
}

func UpdatePassword(c *gin.Context) {
	var input models.UpdatePasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDVal, exists := c.Get("userID")

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	stringUserID, ok := userIDVal.(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid userID format"})
		return
	}

	userObjectID, err := primitive.ObjectIDFromHex(stringUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid userID"})
		return
	}

	usersCollection := db.GetCollection(db.Client, "users")
	var user models.User

	if err := usersCollection.FindOne(context.Background(), bson.M{"_id": userObjectID}).Decode(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "This user doesn't exist"})
	}

	if !utils.CheckPassword(user.Password, input.OldPassword) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Old password is incorrect"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	updateData := bson.M{
		"password":   string(hashedPassword),
		"updated_at": time.Now(),
	}

	result, err := usersCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": userObjectID},
		bson.M{"$set": updateData},
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

}

type CreateAcc struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Role     string `json:"role" binding:"required"`
}

func CreateAccc(c *gin.Context) {
	var input CreateAcc
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userCollection := db.GetCollection(db.Client, "users")
	var existingUser models.User
	err := userCollection.FindOne(context.TODO(), bson.M{"email": input.Email}).Decode(&existingUser)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		return
	}
	
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not hash password"})
		return
	}

	newUser := models.User{
		ID:        primitive.NewObjectID(),
		FirstName: "test Acc",
		LastName:  "test Acc",
		Email:     input.Email,
		Password:  string(hashedPassword),
		Phone:     "test Acc",
		Address:   "test Acc",
		Role:      input.Role,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err = userCollection.InsertOne(context.TODO(), newUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully"})
}
