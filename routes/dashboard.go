package routes

import (
	"github.com/Msaad668/vl-design-backend/controllers"
	"github.com/Msaad668/vl-design-backend/middlewares"
	"github.com/gin-gonic/gin"
)

func RegisterDashboardRoutes(router *gin.Engine) {
	dashboard := router.Group("/dashboard")
	{
		dashboard.GET("designer", middlewares.JWTAuthMiddleware(), middlewares.AllowDesignerOnly(), controllers.GetDesignerDashboard)
	}
}