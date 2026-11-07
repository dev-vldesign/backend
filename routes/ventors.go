package routes

import (
	"github.com/Msaad668/vl-design-backend/controllers"
	"github.com/gin-gonic/gin"
)

func VentorRoutes(router *gin.Engine) {
	// Ventor routes
	router.POST("ventors", controllers.CreateVentor)
	router.GET("ventors", controllers.GetVentors)
	router.GET("ventors/:id", controllers.GetVentorByID)
}
