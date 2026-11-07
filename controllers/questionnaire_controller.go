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
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetQuestionnaireForProject(c *gin.Context) {
	projectID := c.Param("projectID")
	projectObjectID, err := primitive.ObjectIDFromHex(projectID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	collection := db.GetCollection(db.Client, "questionnaires")

	var questionnaire models.Questionnaire
	err = collection.FindOne(context.TODO(), bson.M{"project_id": projectObjectID}).Decode(&questionnaire)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Questionnaire not found", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, questionnaire)
}

func UpdateQuestionAnswer(c *gin.Context) {
	projectID := c.Param("projectID")
	questionID := c.Param("questionID")

	projectObjectID, err := primitive.ObjectIDFromHex(projectID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	questionObjectID, err := primitive.ObjectIDFromHex(questionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid question ID"})
		return
	}

	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid form data"})
		return
	}

	// Handle answer which could be string or string array
	var answer []string
	if answers, exists := c.Request.PostForm["answer"]; exists {
		answer = answers // Always use array even if single value
	}

	// Documents handling (unchanged from your working version)
	documents := []models.Document{}
	files := c.Request.MultipartForm.File["documents"]
	tempDir := "./temp"
	if len(files) > 0 {
		if _, err := os.Stat(tempDir); os.IsNotExist(err) {
			os.MkdirAll(tempDir, os.ModePerm)
		}

		userID, _ := c.Get("userID")
		userIDStr, _ := userID.(string)
		uploadedBy, _ := primitive.ObjectIDFromHex(userIDStr)

		for _, fileHeader := range files {
			filePath := filepath.Join(tempDir, fileHeader.Filename)

			file, err := fileHeader.Open()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open uploaded file"})
				return
			}
			defer file.Close()

			out, err := os.Create(filePath)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save uploaded file"})
				return
			}
			defer out.Close()

			if _, err := io.Copy(out, file); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write file"})
				return
			}

			doc := models.Document{
				ID:         primitive.NewObjectID(),
				Name:       fileHeader.Filename,
				Size:       fmt.Sprintf("%d KB", fileHeader.Size/1024),
				URL:        "/temp/" + fileHeader.Filename,
				UploadedBy: uploadedBy,
			}
			documents = append(documents, doc)
		}
	}

	questionnairesCollection := db.GetCollection(db.Client, "questionnaires")

	filter := bson.M{"project_id": projectObjectID, "questions._id": questionObjectID}

	update := bson.M{
		"$set": bson.M{
			"questions.$.answer":     answer,
			"questions.$.updated_at": time.Now(),
		},
	}

	// Only set answer if it exists
	if answers, exists := c.Request.PostForm["answer"]; exists && len(answers) > 0 {
		update["$set"].(bson.M)["questions.$.answer"] = answers
	}

	// Only set answer if it other_answer exists
	if otherAnswers, exists := c.Request.PostForm["other_answer"]; exists && len(otherAnswers) > 0 && otherAnswers[0] != "" {
		// Take the first element if it's an array, or the single value
		update["$set"].(bson.M)["questions.$.other_answer"] = otherAnswers[0]
	}

	// Handle documents if they exist
	if len(documents) > 0 {
		update["$push"] = bson.M{"questions.$.documents": bson.M{"$each": documents}}
	}

	result, err := questionnairesCollection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update question"})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Question not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Question updated successfully", "result": result})
}
