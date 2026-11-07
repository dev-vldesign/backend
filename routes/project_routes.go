package routes

import (
	"github.com/Msaad668/vl-design-backend/controllers"
	"github.com/Msaad668/vl-design-backend/middlewares"

	"github.com/gin-gonic/gin"
)

func RegisterProjectRoutes(router *gin.Engine) {
    projectGroup := router.Group("/projects")
    {
        projectGroup.GET("/:projectID", middlewares.JWTAuthMiddleware(), controllers.GetProjectByID)
        projectGroup.POST("/", middlewares.JWTAuthMiddleware(), middlewares.AllowCustomerOrAdmin(), controllers.CreateProject)
        projectGroup.GET("/", middlewares.JWTAuthMiddleware(), controllers.GetProjectsForUser)
        projectGroup.GET("/:projectID/questionnaire", middlewares.JWTAuthMiddleware(), controllers.GetQuestionnaireForProject)
        projectGroup.POST("/:projectID/questionnaire/:questionID", middlewares.JWTAuthMiddleware(), controllers.UpdateQuestionAnswer)
        projectGroup.DELETE("/:projectID", controllers.DeleteProject)
    }
}