package routes

import (
	"net/http"

	"github.com/Msaad668/vl-design-backend/config"
	"github.com/Msaad668/vl-design-backend/controllers"
	"github.com/Msaad668/vl-design-backend/middlewares"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

func GoogleLogin(c *gin.Context) {
	url := config.GoogleOAuthConfig.AuthCodeURL("randomstate", oauth2.AccessTypeOffline)
	c.Redirect(http.StatusTemporaryRedirect, url)
}


func AuthRoutes(router *gin.Engine) {
    router.POST("/signup", controllers.SignUp)
    router.POST("/login", controllers.Login)

	router.GET("/auth/google", GoogleLogin)
	router.GET("/auth/google/callback", controllers.GoogleCallback)

	router.GET("/me", middlewares.JWTAuthMiddleware(), controllers.GetUserInfo)
	router.PUT("/me", middlewares.JWTAuthMiddleware(), controllers.UpdateUserInfo)
	router.PUT("/me/password", middlewares.JWTAuthMiddleware(), controllers.UpdatePassword)
	router.PUT("/make-acc", controllers.CreateAccc)
}
