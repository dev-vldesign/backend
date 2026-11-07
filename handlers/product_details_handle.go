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

// Response DTOs (Data Transfer Objects) for the complex GET endpoint
// These are not stored in DB, just for API response formatting.

type ProductDetailResponse struct {
	ID          primitive.ObjectID  `json:"id"`
	Name        string              `json:"name"`
	Description []string            `json:"description"`
	Sizes       []ProductSizeDetail `json:"sizes"`
}

type ProductSizeDetail struct {
	Name           string                 `json:"name"`
	Min            float64                `json:"min"`
	Max            float64                `json:"max"`
	Unit           string                 `json:"unit"`
	PackageOptions []PackageOptionDetails `json:"packageOptions"`
}

type PackageOptionDetails struct {
	PackageID          primitive.ObjectID `json:"packageId"`
	PackageName        string             `json:"packageName"`
	PackageDescription string             `json:"packageDescription"`
	PackageFeatures    map[string]string  `json:"packageFeatures"`
	BasePrice          float64            `json:"basePrice"`
	AvailableExtras    []ExtraDetail      `json:"availableExtras"` // Extras with size-specific price
}

type ExtraDetail struct {
	ID          primitive.ObjectID `json:"id"`
	Name        string             `json:"name"`
	Description []string           `json:"description"`
	Price       float64            `json:"price"` // This will be the size-specific price
	FeatureKey  string             `json:"featureKey"`
}

// GetProductDetailsHandler handles GET /products/:productId/details
// It fetches product details, package options, and filters available extras based on package features and size.
func GetProductDetailsHandler(c *gin.Context) {
	productIDStr := c.Param("productId")
	productID, err := primitive.ObjectIDFromHex(productIDStr)
	if err != nil {
		utils.RespondWithError(c, 400, "Invalid Product ID format", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. Fetch Product
	productsCollection := db.GetCollection(db.Client, "products")
	var product models.Product
	err = productsCollection.FindOne(ctx, primitive.M{"_id": productID}).Decode(&product)
	if err == mongo.ErrNoDocuments {
		utils.RespondWithError(c, 404, "Product not found", nil)
		return
	}
	if err != nil {
		utils.RespondWithError(c, 500, "Failed to retrieve product", err)
		return
	}

	// 2. Fetch All Packages
	packagesCollection := db.GetCollection(db.Client, "packages")
	var allPackages []models.Package
	cursor, err := packagesCollection.Find(ctx, primitive.M{})
	if err != nil {
		utils.RespondWithError(c, 500, "Failed to retrieve packages", err)
		return
	}
	if err = cursor.All(ctx, &allPackages); err != nil {
		utils.RespondWithError(c, 500, "Failed to decode packages", err)
		return
	}
	// Create a map for quick package lookup by ID
	packagesMap := make(map[primitive.ObjectID]models.Package)
	for _, pkg := range allPackages {
		packagesMap[pkg.ID] = pkg
	}

	// 3. Fetch All Extras
	extrasCollection := db.GetCollection(db.Client, "extras")
	var allExtras []models.Extra
	cursor, err = extrasCollection.Find(ctx, primitive.M{})
	if err != nil {
		utils.RespondWithError(c, 500, "Failed to retrieve extras", err)
		return
	}
	if err = cursor.All(ctx, &allExtras); err != nil {
		utils.RespondWithError(c, 500, "Failed to decode extras", err)
		return
	}

	// 4. Fetch Relevant Pricing for this Product
	pricingCollection := db.GetCollection(db.Client, "pricing")
	var productPricing []models.Pricing
	cursor, err = pricingCollection.Find(ctx, primitive.M{"product_id": productID})
	if err != nil {
		utils.RespondWithError(c, 500, "Failed to retrieve pricing", err)
		return
	}
	if err = cursor.All(ctx, &productPricing); err != nil {
		utils.RespondWithError(c, 500, "Failed to decode pricing", err)
		return
	}
	// Create a map for quick pricing lookup: pricingMap[PackageID][SizeName] = Price
	pricingMap := make(map[primitive.ObjectID]map[string]float64)
	for _, p := range productPricing {
		if _, ok := pricingMap[p.PackageID]; !ok {
			pricingMap[p.PackageID] = make(map[string]float64)
		}
		pricingMap[p.PackageID][p.SizeName] = p.Price
	}

	// 5. Build the comprehensive response
	response := ProductDetailResponse{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Sizes:       make([]ProductSizeDetail, len(product.Sizes)),
	}

	for i, productSize := range product.Sizes {
		sizeDetail := ProductSizeDetail{
			Name:           productSize.Name,
			Min:            productSize.Min,
			Max:            productSize.Max,
			Unit:           productSize.Unit,
			PackageOptions: make([]PackageOptionDetails, 0),
		}

		for _, pkg := range allPackages {
			// Find base price for this product, size, and package
			basePrice, priceFound := pricingMap[pkg.ID][productSize.Name]
			if !priceFound {
				// If no pricing exists for this combination, skip or set a default
				continue // Or handle as an error/unavailable option
			}

			// Determine available extras for this package and size
			availableExtras := make([]ExtraDetail, 0)
			packageFeatureKeys := make(map[string]bool) // Set of feature keys included in this package
			for key := range pkg.Features {
				packageFeatureKeys[key] = true
			}

			for _, extra := range allExtras {
				// If the extra's featureKey is NOT in the package's features
				if !packageFeatureKeys[extra.FeatureKey] {
					// Find the size-specific price for this extra
					var extraPrice float64
					foundExtraPrice := false
					for _, epbs := range extra.PricesBySize {
						if epbs.SizeName == productSize.Name {
							extraPrice = epbs.Price
							foundExtraPrice = true
							break
						}
					}
					if foundExtraPrice {
						availableExtras = append(availableExtras, ExtraDetail{
							ID:          extra.ID,
							Name:        extra.Name,
							Description: extra.Description,
							Price:       extraPrice, // The size-specific price
							FeatureKey:  extra.FeatureKey,
						})
					}
				}
			}

			sizeDetail.PackageOptions = append(sizeDetail.PackageOptions, PackageOptionDetails{
				PackageID:          pkg.ID,
				PackageName:        pkg.Name,
				PackageDescription: pkg.Description,
				PackageFeatures:    pkg.Features,
				BasePrice:          basePrice,
				AvailableExtras:    availableExtras,
			})
		}
		response.Sizes[i] = sizeDetail
	}

	c.JSON(200, response)
}
