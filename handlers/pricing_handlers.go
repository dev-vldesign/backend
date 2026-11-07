package handlers

import (
	"context"
	"time"

	"github.com/Msaad668/vl-design-backend/db"
	"github.com/Msaad668/vl-design-backend/models"
	"github.com/Msaad668/vl-design-backend/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CreatePricingHandler handles POST /pricing to set a base price for a Product-Package-Size combo.
func CreatePricingHandler(c *gin.Context) {
	var pricing models.Pricing
	if err := c.ShouldBindJSON(&pricing); err != nil {
		utils.RespondWithError(c, 400, "Invalid pricing data", err)
		return
	}

	// Optional: Validate if ProductID and PackageID exist and SizeName is valid for Product
	// This would involve additional DB lookups.

	pricing.ID = primitive.NewObjectID()
	pricing.CreatedAt = time.Now()
	pricing.UpdatedAt = time.Now()

	collection := db.GetCollection(db.Client, "pricing")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, pricing)
	if err != nil {
		utils.RespondWithError(c, 500, "Failed to create pricing entry", err)
		return
	}

	c.JSON(201, pricing) // 201 Created
}
