package routes

import (
	"github.com/dat1010/go-api/controllers"
	"github.com/dat1010/go-api/middleware"
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

	api.GET("", controllers.ListPosts)
	api.GET("/:id", controllers.GetPost)

	// Protected routes
	api.Use(middleware.Auth0())
	api.POST("", controllers.CreatePost)
	api.PUT("/:id", controllers.UpdatePost)
	api.DELETE("/:id", controllers.DeletePost)
}
