package routes

import (
	"github.com/Msaad668/vl-design-backend/controllers"
	"github.com/Msaad668/vl-design-backend/middlewares"
	"github.com/gin-gonic/gin"
)

func RegisterDesinerPortfolioRoutes(r *gin.Engine) {
	portfolioRoutes := r.Group("/portfolio")
	portfolioRoutes.Use(middlewares.JWTAuthMiddleware(), middlewares.AllowDesignerOnly())
	{
		portfolioRoutes.GET("", controllers.GetPortfolio)
		portfolioRoutes.PUT("", controllers.UpdatePortfolio)
		portfolioRoutes.PUT("image", controllers.UpdatePortfolioImage)

		portfolioRoutes.POST("project", controllers.AddProjectToPortfolio)
		portfolioRoutes.DELETE("project/:projectID", controllers.DeleteExistingProject)
		portfolioRoutes.PUT("project/:projectID", controllers.AddImageToProject)
		portfolioRoutes.PUT("project/:projectID/:imageID", controllers.DeleteImageFromProject)
	}
}
