package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/Msaad668/vl-design-backend/db"
	"github.com/Msaad668/vl-design-backend/models"
	"github.com/Msaad668/vl-design-backend/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TutorialController struct {
	Collection *mongo.Collection
}

func NewTutorialController(collection *mongo.Collection) *TutorialController {
	return &TutorialController{Collection: collection}
}

func CreateTutorial(c *gin.Context) {
	TutorialCollection := db.GetCollection(db.Client, "tutorials")

	now := time.Now()

	title := c.PostForm("title")
	description := c.PostForm("description")
	sectionsJson := c.PostForm("sections")

	if title == "" || sectionsJson == "" {
		c.JSON(400, gin.H{"error": "Missing title or sections"})
		return
	}

	type IncomingSection struct {
		Title     string   `json:"title"`
		VideoKeys []string `json:"videoKeys"`
	}

	var incomingSections []IncomingSection
	if err := json.Unmarshal([]byte(sectionsJson), &incomingSections); err != nil {
		c.JSON(400, gin.H{"error": "Invalid sections JSON", "details": err.Error()})
		return
	}

	// Ensure upload dir exists
	if err := os.MkdirAll("uploads", os.ModePerm); err != nil {
		c.JSON(500, gin.H{"error": "Failed to create uploads folder"})
		return
	}

	tutorial := models.Tutorial{
		ID:          primitive.NewObjectID(),
		Title:       title,
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
		Sections:    []models.Section{},
	}

	for idx, s := range incomingSections {
		var lessons []models.Lesson
		var sectionDuration time.Duration = 0

		for _, videoKey := range s.VideoKeys {
			file, err := c.FormFile(videoKey)
			if err != nil {
				c.JSON(400, gin.H{"error": fmt.Sprintf("Missing video: %s", videoKey)})
				return
			}

			filename := fmt.Sprintf("uploads/%d_%s", time.Now().UnixNano(), file.Filename)
			if err := c.SaveUploadedFile(file, filename); err != nil {
				c.JSON(500, gin.H{"error": "Failed to save file", "details": err.Error()})
				return
			}

			durationSec, err := utils.GetVideoDuration(filename)
			if err != nil {
				c.JSON(500, gin.H{"error": "Failed to get video duration", "details": err.Error()})
				return
			}

			duration := time.Duration(durationSec * float64(time.Second))
			sectionDuration += duration

			lessons = append(lessons, models.Lesson{
				ID:        primitive.NewObjectID(),
				Title:     file.Filename,
				Url:       filename,
				Duration:  duration,
				CreatedAt: now,
				UpdatedAt: now,
			})
		}

		tutorial.Sections = append(tutorial.Sections, models.Section{
			ID:        primitive.NewObjectID(),
			Title:     s.Title,
			Order:     idx + 1,
			Lessons:   lessons,
			Duration:  sectionDuration,
			CreatedAt: now,
			UpdatedAt: now,
		})
	}

	if _, err := TutorialCollection.InsertOne(context.Background(), tutorial); err != nil {
		c.JSON(500, gin.H{"error": "Failed to insert tutorial", "details": err.Error()})
		return
	}

	c.JSON(201, gin.H{"message": "Tutorial created successfully", "data": tutorial})
}

func GetTutorials(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "6")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		pageSize = 6
	}

	skip := (page - 1) * pageSize

	// ✅ use the collection from db utility
	collection := db.GetCollection(db.Client, "tutorials")

	totalCount, err := collection.CountDocuments(context.Background(), bson.M{})

	if err != nil {
		return
	}
	
	// Projection
	projection := bson.M{
		"title":       1,
		"description": 1,
		"thumbnail":   1,
	}

	// ✅ applying options
	cursor, err := collection.Find(
		context.Background(),
		bson.M{}, // empty filter 
		options.Find().
			SetProjection(projection).
			SetSkip(int64(skip)).
			SetLimit(int64(pageSize)).
			SetSort(bson.M{"created_at": -1}),
	)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch tutorials", "details": err.Error()})
		return
	}
	defer cursor.Close(context.Background())

	var tutorials []bson.M
	if err := cursor.All(context.Background(), &tutorials); err != nil {
		c.JSON(500, gin.H{"error": "Failed to decode tutorials", "details": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"page":     page,
		"pageSize": pageSize,
		"items":    tutorials,
		"totalCount": totalCount,
	})
}

func GetTutorialByID(c *gin.Context) {
	id := c.Param("id")

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid ID format"})
		return
	}

	collection := db.GetCollection(db.Client, "tutorials")

	var tutorial bson.M
	err = collection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&tutorial)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(404, gin.H{"error": "Tutorial not found"})
		} else {
			c.JSON(500, gin.H{"error": "Failed to fetch tutorial", "details": err.Error()})
		}
		return
	}

	c.JSON(200, tutorial)
}
