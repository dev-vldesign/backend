package main

import (
	"time"

	"github.com/Msaad668/vl-design-backend/config"
	"github.com/Msaad668/vl-design-backend/db"
	"github.com/Msaad668/vl-design-backend/routes"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load environment variables
	config.LoadEnv()

	config.SetupGoogleOAuth()

	// Connect to MongoDB
	db.ConnectDB()

	// Ensure indexes
	db.EnsureIndexes(db.Client)

	// Initialize Gin router
	r := gin.Default()

	// Configure CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome to vl-design-backend!",
		})
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "OK",
		})
	})

	// serve static files
	r.Static("/temp", "./temp")
	r.Static("/uploads", "./uploads")

	routes.AuthRoutes(r)
	routes.RegisterProjectRoutes(r)
	routes.RegisterthreadsRoutes(r)
	routes.RegisterNotificationRoutes(r)
	routes.AdminRoutes(r)
	routes.PaymentRoutes(r)
	routes.RegisterTemplatesRoutes(r)
	routes.RegisterTutorialsRoutes(r)
	routes.RegisterDashboardRoutes(r)
	routes.RegisterDesinerPortfolioRoutes(r)
	routes.VentorRoutes(r)

	r.Run(":8080")
}
