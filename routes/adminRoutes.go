package routes

import (
	"github.com/Msaad668/vl-design-backend/controllers"
	"github.com/Msaad668/vl-design-backend/middlewares"
	"github.com/gin-gonic/gin"
)

func AdminRoutes(router *gin.Engine) {
	adminGroup := router.Group("/admin")
	adminGroup.Use(middlewares.JWTAuthMiddleware(), middlewares.AllowAdminOnly())
	{
		adminGroup.PUT("/projects/:projectID", controllers.AdminUpdateProject)
		adminGroup.PUT("/projects/:projectID/tasks/:taskID", controllers.AdminUpdateProjectTask)
		adminGroup.GET("/users", controllers.GetAllUsers)
	}
}
