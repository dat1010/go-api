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
	api.POST("/refresh", controllers.Refresh)

	// Public routes
	api.GET("/healthcheck", controllers.GetHealthCheck)
	api.GET("/secrets", controllers.GetSecret)
	api.GET("/discord-ping", controllers.PingDiscord)

	// Protected routes
	protected := api.Group("")
	protected.Use(middleware.Auth0())
	protected.Use(middleware.EnsureUserRole("member"))
	protected.GET("/me", controllers.CheckAuth)
	protected.POST("/events", controllers.CreateEvent)
	protected.GET("/events", controllers.ListUserEvents)

	// Admin routes (superadmin only)
	admin := api.Group("/admin")
	admin.Use(middleware.Auth0())
	admin.Use(middleware.EnsureUserRole("member"))
	admin.Use(middleware.RequireRole("superadmin"))
	admin.GET("/users", controllers.ListUsers)
	admin.POST("/users", controllers.CreateUser)
	admin.PATCH("/users/:id/role", controllers.UpdateUserRole)
	admin.DELETE("/users/:id", controllers.DeleteUser)
}
