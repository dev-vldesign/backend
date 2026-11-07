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

// CreatePackageHandler handles POST /packages to create a new package.
func CreatePackageHandler(c *gin.Context) {
	var pkg models.Package
	if err := c.ShouldBindJSON(&pkg); err != nil {
		utils.RespondWithError(c, 400, "Invalid package data", err)
		return
	}

	pkg.ID = primitive.NewObjectID()
	pkg.CreatedAt = time.Now()
	pkg.UpdatedAt = time.Now()

	collection := db.GetCollection(db.Client, "packages")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, pkg)
	if err != nil {
		utils.RespondWithError(c, 500, "Failed to create package", err)
		return
	}

	c.JSON(201, pkg) // 201 Created
}

// GetPackageHandler handles GET /packages/:packageId to retrieve a single package.
func GetPackageHandler(c *gin.Context) {
	packageIDStr := c.Param("packageId")
	packageID, err := primitive.ObjectIDFromHex(packageIDStr)
	if err != nil {
		utils.RespondWithError(c, 400, "Invalid Package ID format", err)
		return
	}

	collection := db.GetCollection(db.Client, "packages")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var pkg models.Package
	err = collection.FindOne(ctx, primitive.M{"_id": packageID}).Decode(&pkg)
	if err == mongo.ErrNoDocuments {
		utils.RespondWithError(c, 404, "Package not found", nil)
		return
	}
	if err != nil {
		utils.RespondWithError(c, 500, "Failed to retrieve package", err)
		return
	}

	c.JSON(200, pkg)
}
