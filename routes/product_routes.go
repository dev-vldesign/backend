package routes

import (
	"github.com/Msaad668/vl-design-backend/handlers"
	"github.com/Msaad668/vl-design-backend/middlewares"
	"github.com/gin-gonic/gin"
)

func ProductRoutes(router *gin.Engine) {

	// Product Endpoints
	// Public GET for product details
	router.GET("/product/:productId/details", handlers.GetProductDetailsHandler)

	// Package Endpoints
	// Public GET for a specific package
	router.GET("/package/:packageId", handlers.GetPackageHandler)

	// Extra Endpoints
	// Public GET for a specific extra (if you want to allow fetching individual extra details publicly)
	router.GET("/extra/:extraId", handlers.GetExtraHandler)

	// Admin Protected Routes
	// Group routes that require both JWT authentication and Admin authorization
	adminGroup := router.Group("/") // Base path for admin routes
	adminGroup.Use(middlewares.JWTAuthMiddleware(), middlewares.AllowAdminOnly())
	{
		// Products (Admin Only)
		adminGroup.POST("/product", handlers.CreateProductHandler)
		// adminGroup.PUT("/products/:productId", handlers.UpdateProductHandler) // Example update
		// adminGroup.DELETE("/products/:productId", handlers.DeleteProductHandler) // Example delete

		// Packages (Admin Only)
		adminGroup.POST("/package", handlers.CreatePackageHandler)
		// adminGroup.PUT("/packages/:packageId", handlers.UpdatePackageHandler) // Example update
		// adminGroup.DELETE("/packages/:packageId", handlers.DeletePackageHandler) // Example delete

		// Extras (Admin Only)
		adminGroup.POST("/extra", handlers.CreateExtraHandler) // Corrected handler!
		// adminGroup.PUT("/extras/:extraId", handlers.UpdateExtraHandler) // Example update
		// adminGroup.DELETE("/extras/:extraId", handlers.DeleteExtraHandler) // Example delete

		// Pricing (Admin Only)
		adminGroup.POST("/pricing", handlers.CreatePricingHandler)
		// adminGroup.PUT("/pricing/:pricingId", handlers.UpdatePricingHandler) // Example update
		// adminGroup.DELETE("/pricing/:pricingId", handlers.DeletePricingHandler) // Example delete
	}

}
