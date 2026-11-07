package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/Msaad668/vl-design-backend/db"
	"github.com/Msaad668/vl-design-backend/models"
	"github.com/Msaad668/vl-design-backend/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// CreateVentor creates a new ventor in the database
func CreateVentor(c *gin.Context) {
	var input models.CreateVentorInput

	// Bind the form data to the input struct
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid input data",
			"details": err.Error(),
		})
		return
	}

	// Validate input using utility function
	if err := utils.ValidateVentorInput(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Check if ventor with same email already exists
	ventorCollection := db.GetCollection(db.Client, "ventors")

	var existingVentor models.Ventor
	err := ventorCollection.FindOne(context.Background(), map[string]interface{}{
		"email": input.Email,
	}).Decode(&existingVentor)

	if err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"error": "A ventor with this email already exists",
		})
		return
	} else if err != mongo.ErrNoDocuments {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error checking existing ventor",
		})
		return
	}

	// Handle file uploads using utility function
	logoPath, licensePath, err := utils.HandleFileUploads(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Create ventor using utility function
	ventor := utils.CreateVentorFromInput(&input, logoPath, licensePath)
	ventor.CreatedAt = time.Now()
	ventor.UpdatedAt = time.Now()

	// Insert the ventor into the database
	_, err = ventorCollection.InsertOne(context.Background(), ventor)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create ventor",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Ventor created successfully",
		"ventor":  ventor,
	})
}

// GetVentors retrieves all ventors from the database
func GetVentors(c *gin.Context) {
	ventorCollection := db.GetCollection(db.Client, "ventors")

	cursor, err := ventorCollection.Find(context.Background(), map[string]interface{}{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch ventors",
		})
		return
	}
	defer cursor.Close(context.Background())

	var ventors []models.Ventor
	if err := cursor.All(context.Background(), &ventors); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to decode ventors",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ventors": ventors,
		"count":   len(ventors),
	})
}

// GetVentorByID retrieves a specific ventor by ID
func GetVentorByID(c *gin.Context) {
	ventorID := c.Param("id")

	objectID, err := primitive.ObjectIDFromHex(ventorID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid ventor ID format",
		})
		return
	}

	ventorCollection := db.GetCollection(db.Client, "ventors")

	var ventor models.Ventor
	err = ventorCollection.FindOne(context.Background(), map[string]interface{}{
		"_id": objectID,
	}).Decode(&ventor)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Ventor not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch ventor",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ventor": ventor,
	})
}
