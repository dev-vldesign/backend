package controllers

import (
	"context"
	"net/http"

	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Msaad668/vl-design-backend/db"
	"github.com/Msaad668/vl-design-backend/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// handlers/portfolio.go
func GetPortfolio(c *gin.Context) {
	// 1. Extract and validate user ID from context
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	userIDStr, ok := userIDVal.(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	userObjectID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// 2. Get portfolio collection
	collection := db.GetCollection(db.Client, "portfolios")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 3. Find portfolio by user ID
	var portfolio models.Portfolio
	err = collection.FindOne(ctx, bson.M{"user_id": userObjectID}).Decode(&portfolio)

	// 4. Handle case where portfolio doesn't exist by creating one
	if err == mongo.ErrNoDocuments {
		// Get user info to create portfolio
		user, err := fetchUserByID(userObjectID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user info"})
			return
		}

		// Create new portfolio with user's name
		portfolio = models.Portfolio{
			ID:        primitive.NewObjectID(),
			UserID:    userObjectID,
			Name:      user.FirstName + " " + user.LastName,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Projects:  []models.PortfolioProject{},
			Tools:     []string{},
		}

		// Insert the new portfolio
		_, err = collection.InsertOne(ctx, portfolio)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create portfolio"})
			return
		}

		// Return the newly created portfolio
		c.JSON(http.StatusOK, gin.H{
			"portfolio": portfolio,
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch portfolio"})
		return
	}

	// 5. Return the portfolio data
	c.JSON(http.StatusOK, gin.H{
		"portfolio": portfolio,
	})
}

func UpdatePortfolio(c *gin.Context) {
	userIDrow, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userIDStr, ok := userIDrow.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form"})
		return
	}

	form := c.Request.MultipartForm
	col := db.GetCollection(db.Client, "portfolios")

	// Fetch existing portfolio
	var existingPortfolio models.Portfolio
	err = col.FindOne(context.TODO(), bson.M{"user_id": userID}).Decode(&existingPortfolio)
	if err == mongo.ErrNoDocuments {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Portfolio not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch portfolio"})
		return
	}

	// --- Prepare update fields ---
	update := bson.M{}
	if name := getFirst(form.Value["name"]); name != "" {
		update["name"] = name
	}
	if about := getFirst(form.Value["about"]); about != "" {
		update["about"] = about
	}
	if experience := getFirst(form.Value["experience"]); experience != "" {
		update["experience"] = experience
	}
	if nationality := getFirst(form.Value["nationality"]); nationality != "" {
		update["nationality"] = nationality
	}
	if tools := getFirst(form.Value["tools"]); tools != "" {
		update["tools"] = strings.Split(tools, ",")
	}

	update["updated_at"] = time.Now()

	_, err = col.UpdateOne(
		context.TODO(),
		bson.M{"user_id": userID},
		bson.M{"$set": update},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update portfolio"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Portfolio updated successfully"})
}

func getFirst(values []string) string {
	if len(values) > 0 {
		return values[0]
	}
	return ""
}

func UpdatePortfolioImage(c *gin.Context) {
	// Get and validate userID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	designerID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get existing portfolio
	col := db.GetCollection(db.Client, "portfolios")
	var existingPortfolio models.Portfolio
	err = col.FindOne(context.TODO(), bson.M{"user_id": designerID}).Decode(&existingPortfolio)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Portfolio not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch portfolio"})
		}
		return
	}

	// Get the file from form
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No image file provided"})
		return
	}
	defer file.Close()

	// Validate image type
	if !strings.Contains(header.Header.Get("Content-Type"), "image/") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only images are allowed"})
		return
	}

	// Determine which field to update based on the form key
	imageType := c.PostForm("type")
	if imageType != "profile_picture" && imageType != "cover" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid image type. Must be 'profile_picture' or 'cover'"})
		return
	}

	// Save new image
	filename := fmt.Sprintf("%d_%s_%s", time.Now().UnixNano(), imageType, header.Filename)
	savePath := filepath.ToSlash(filepath.Join("uploads", "portfolios", filename))

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Join("uploads", "portfolios"), os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
		return
	}

	// Create the file
	out, err := os.Create(savePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image", "message": err.Error()})
		return
	}
	defer out.Close()

	// Copy the uploaded file to the destination
	if _, err := io.Copy(out, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image", "message": err.Error()})
		return
	}

	// Delete old image if it exists
	var oldImagePath string
	updateField := ""

	if imageType == "profile_picture" {
		oldImagePath = existingPortfolio.ProfilePicture
		updateField = "profile_picture"
	} else {
		oldImagePath = existingPortfolio.Cover
		updateField = "cover"
	}

	if oldImagePath != "" && strings.HasPrefix(oldImagePath, "/") {
		if err := os.Remove(filepath.Join(".", oldImagePath)); err != nil {
			// Log error but continue, as this is not critical
			fmt.Printf("Failed to delete old image: %v\n", err)
		}
	}

	// Update the portfolio with new image path
	newImagePath := "/" + savePath
	update := bson.M{
		"$set": bson.M{
			updateField:  newImagePath,
			"updated_at": time.Now(),
		},
	}

	fmt.Println("\n update : ", update)
	_, err = col.UpdateOne(context.TODO(), bson.M{"user_id": designerID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update portfolio"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("%s updated successfully", imageType),
		"path":    newImagePath,
	})
}

func AddProjectToPortfolio(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	userIDStr, ok := userIDVal.(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	userObjectID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	portfolioCollection := db.GetCollection(db.Client, "portfolios")

	var userPortfolio models.Portfolio
	err = portfolioCollection.FindOne(c, bson.M{"user_id": userObjectID}).Decode(&userPortfolio)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Can't find a portfolio for this user"})
	}

	title := c.PostForm("title")
	if title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title is required"})
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form data"})
		return
	}

	files := form.File["images"]

	// Save images and collect URLs

	var images []models.Image
	for _, file := range files {
		imageID := primitive.NewObjectID()
		filename := fmt.Sprintf("%s_%s", primitive.NewObjectID().Hex(), file.Filename)
		savePath := filepath.ToSlash(filepath.Join("uploads", "portfolios", filename))

		// Save the file
		if err := c.SaveUploadedFile(file, savePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
			return
		}

		// Optional: add your base URL if needed
		images = append(images, models.Image{
			ID:   imageID,
			Path: savePath,
		})
	}

	newProject := models.PortfolioProject{
		ID:     primitive.NewObjectID(),
		Title:  title,
		Images: images,
	}

	// Append the project to the existing portfolio
	update := bson.M{
		"$push": bson.M{"projects": newProject},
		"$set":  bson.M{"updated_at": time.Now()},
	}

	_, err = portfolioCollection.UpdateOne(c, bson.M{"user_id": userObjectID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update portfolio"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Project added successfully", "project": newProject})

}

func DeleteExistingProject(c *gin.Context) {

	projectIDStr := c.Param("projectID")
	projectID, err := primitive.ObjectIDFromHex(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	userIDVal, exists := c.Get("userID")

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	userIDStr, ok := userIDVal.(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	userObjectID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	portfolioCollection := db.GetCollection(db.Client, "portfolios")

	var userPortfolio models.Portfolio
	err = portfolioCollection.FindOne(c, bson.M{"user_id": userObjectID}).Decode(&userPortfolio)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Can't find a portfolio for this user"})
	}

	projectExists := false
	for _, project := range userPortfolio.Projects {
		if project.ID == projectID {
			projectExists = true
			break
		}
	}

	if !projectExists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found in user's portfolio"})
		return
	}

	update := bson.M{
		"$pull": bson.M{
			"projects": bson.M{"_id": projectID},
		},
		"$set": bson.M{"updated_at": time.Now()},
	}

	result, err := portfolioCollection.UpdateOne(c, bson.M{"user_id": userObjectID}, update)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove project"})
		return
	}

	if result.ModifiedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found or already deleted"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Project deleted successfully"})

}

func AddImageToProject(c *gin.Context) {
	projectIDStr := c.Param("projectID")
	projectID, err := primitive.ObjectIDFromHex(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	userIDVal, exists := c.Get("userID")

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	userIDStr, ok := userIDVal.(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	userObjectID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	portfolioCollection := db.GetCollection(db.Client, "portfolios")

	var userPortfolio models.Portfolio
	err = portfolioCollection.FindOne(c, bson.M{"user_id": userObjectID}).Decode(&userPortfolio)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Can't find a portfolio for this user"})
	}

	projectExists := false
	for _, project := range userPortfolio.Projects {
		if project.ID == projectID {
			projectExists = true
			break
		}
	}

	if !projectExists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found in user's portfolio"})
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form data"})
		return
	}

	files := form.File["images"]

	var images []models.Image
	for _, file := range files {
		imageID := primitive.NewObjectID()
		filename := fmt.Sprintf("%s_%s", primitive.NewObjectID().Hex(), file.Filename)
		savePath := filepath.ToSlash(filepath.Join("uploads", "portfolios", filename))

		// Save the file
		if err := c.SaveUploadedFile(file, savePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
			return
		}

		// Optional: add your base URL if needed
		images = append(images, models.Image{
			ID:   imageID,
			Path: savePath,
		})
	}

	// Push all images in one operation
	filter := bson.M{
		"user_id":      userObjectID,
		"projects._id": projectID,
	}

	update := bson.M{
		"$push": bson.M{
			"projects.$.images": bson.M{
				"$each": images,
			},
		},
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	_, err = portfolioCollection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update portfolio", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Images are added to the project successfully"})
}

func DeleteImageFromProject(c *gin.Context) {
	projectIDStr := c.Param("projectID")
	imageIDStr := c.Param("imageID")

	projectID, err := primitive.ObjectIDFromHex(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	imageID, err := primitive.ObjectIDFromHex(imageIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	userIDVal, exists := c.Get("userID")

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	userIDStr, ok := userIDVal.(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	userObjectID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	portfolioCollection := db.GetCollection(db.Client, "portfolios")

	var userPortfolio models.Portfolio
	err = portfolioCollection.FindOne(c, bson.M{"user_id": userObjectID}).Decode(&userPortfolio)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Can't find a portfolio for this user"})
	}

	projectExists := false
	for _, project := range userPortfolio.Projects {
		if project.ID == projectID {
			projectExists = true
			break
		}
	}

	if !projectExists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found in user's portfolio"})
		return
	}

	filter := bson.M{
		"user_id": userObjectID,
		"projects._id": projectID,
	}

	update := bson.M{
		"$pull" : bson.M{
			"projects.$.images" : bson.M{"_id" : imageID},
		},
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	_, err = portfolioCollection.UpdateOne(context.TODO(), filter, update)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "an error happened while adding the image", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Image deleted successfully"})
}