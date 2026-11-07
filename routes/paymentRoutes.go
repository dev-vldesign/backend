package routes

import (
	"github.com/Msaad668/vl-design-backend/controllers"
	"github.com/gin-gonic/gin"
)

func PaymentRoutes(router *gin.Engine) {
    paymentGroup := router.Group("/checkout")
    {
        paymentGroup.POST("/payments/create-checkout-url", controllers.CreateCheckoutUrl)
    }
}
