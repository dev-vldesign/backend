package routes

import (
	"github.com/Msaad668/vl-design-backend/controllers"
	"github.com/Msaad668/vl-design-backend/middlewares"

	"github.com/gin-gonic/gin"
)

func RegisterthreadsRoutes(router *gin.Engine) {
    projectThreadsGroup := router.Group("/projects/:projectID/threads")
    {
        projectThreadsGroup.POST("/", middlewares.JWTAuthMiddleware(), controllers.AddThread)
        projectThreadsGroup.GET("/", middlewares.JWTAuthMiddleware(), controllers.GetThreadsForProject)
        
    }

    threadRoute := router.Group("/threads/:threadID")
    {
        threadRoute.DELETE("/", middlewares.JWTAuthMiddleware(), controllers.DeleteThread)
        threadRoute.POST("/replies", middlewares.JWTAuthMiddleware(), controllers.AddReplyToThread)
        threadRoute.GET("/replies", middlewares.JWTAuthMiddleware(), controllers.GetRepliesForThread) 
        threadRoute.DELETE("/replies/:replyID", middlewares.JWTAuthMiddleware(), controllers.DeleteReplyFromThread)

        threadRoute.GET("", middlewares.JWTAuthMiddleware(), controllers.GetThreadByID)
    }
}