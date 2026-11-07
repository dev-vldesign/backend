package routes

import (
	"github.com/Msaad668/vl-design-backend/controllers"
	"github.com/Msaad668/vl-design-backend/middlewares"
	"github.com/gin-gonic/gin"
)

func RegisterNotificationRoutes(router *gin.Engine) {
    notificationsGroup := router.Group("/notifications")
    {
        notificationsGroup.GET("", middlewares.JWTAuthMiddleware(), controllers.GetNotificationsForUser)
        notificationsGroup.POST("/mark-seen", middlewares.JWTAuthMiddleware(), controllers.MarkNotificationsAsSeen)
    }
}
