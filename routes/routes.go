package routes

import (
	"github.com/dat1010/go-api/controllers"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers all the API routes
func RegisterRoutes(api *gin.RouterGroup) {
	// Auth routes
	api.GET("/login", controllers.Login)
	api.GET("/callback", controllers.Callback)
	api.GET("/logout", controllers.Logout)

	// Other routes...
	api.GET("/healthcheck", controllers.GetHealthCheck)
	api.GET("/secrets", controllers.GetSecret)
}
