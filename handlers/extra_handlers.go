package handlers

import (
	"context"
	"time"

	"github.com/Msaad668/vl-design-backend/db"
	"github.com/Msaad668/vl-design-backend/models"
	"github.com/Msaad668/vl-design-backend/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// CreateExtraHandler handles POST /extras to create a new extra.
func CreateExtraHandler(c *gin.Context) {
	var extra models.Extra
	if err := c.ShouldBindJSON(&extra); err != nil {
		utils.RespondWithError(c, 400, "Invalid extra data", err)
		return
	}

	extra.ID = primitive.NewObjectID()
	extra.CreatedAt = time.Now()
	extra.UpdatedAt = time.Now()

	collection := db.GetCollection(db.Client, "extras")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, extra)
	if err != nil {
		utils.RespondWithError(c, 500, "Failed to create extra", err)
		return
	}

	c.JSON(201, extra) // 201 Created
}

// GetExtraHandler handles GET /extras/:extraId to retrieve a single extra.
func GetExtraHandler(c *gin.Context) {
	extraIDStr := c.Param("extraId")
	extraID, err := primitive.ObjectIDFromHex(extraIDStr)
	if err != nil {
		utils.RespondWithError(c, 400, "Invalid Extra ID format", err)
		return
	}

	collection := db.GetCollection(db.Client, "extras")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var extra models.Extra
	err = collection.FindOne(ctx, primitive.M{"_id": extraID}).Decode(&extra)
	if err == mongo.ErrNoDocuments {
		utils.RespondWithError(c, 404, "Extra not found", nil)
		return
	}
	if err != nil {
		utils.RespondWithError(c, 500, "Failed to retrieve extra", err)
		return
	}

	c.JSON(200, extra)
}
