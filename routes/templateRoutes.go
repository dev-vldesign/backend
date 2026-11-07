package routes

import (
	"github.com/Msaad668/vl-design-backend/controllers"
	"github.com/Msaad668/vl-design-backend/middlewares"
	"github.com/gin-gonic/gin"
)

func RegisterTemplatesRoutes(router *gin.Engine) {
    templates := router.Group("/templates")
    templates.Use(
        middlewares.JWTAuthMiddleware(),
    )
    
    // Routes accessible to designers or admins
    designerOrAdmin := templates.Group("")
    designerOrAdmin.Use(middlewares.AllowDesignerOrAdmin())
    {
        designerOrAdmin.GET("", controllers.GetTemplates)
        designerOrAdmin.GET(":id", controllers.GetFolderDetails)
    }
    
    // Routes accessible only to admins
    adminOnly := templates.Group("")
    adminOnly.Use(middlewares.AllowAdminOnly())
    {
        adminOnly.POST("", controllers.CreateFolder)
        adminOnly.POST(":id", controllers.UploadFilesToFolder)
        adminOnly.PUT(":id", controllers.UpdateFolderName)
        adminOnly.DELETE(":id", controllers.DeleteFolder)
    }
}