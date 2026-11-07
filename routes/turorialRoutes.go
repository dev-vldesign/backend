package routes

import (
	"github.com/Msaad668/vl-design-backend/controllers"

	"github.com/Msaad668/vl-design-backend/middlewares"
	"github.com/gin-gonic/gin"
)

func RegisterTutorialsRoutes(router *gin.Engine) {
	tutorial := router.Group("/tutorials")
	tutorial.Use(middlewares.JWTAuthMiddleware())
	{
		tutorial.POST("", middlewares.AllowAdminOnly(), controllers.CreateTutorial)
		tutorial.GET("", controllers.GetTutorials)
		tutorial.GET("/:id", controllers.GetTutorialByID)
	}
}
