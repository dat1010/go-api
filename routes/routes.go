package routes

import (
	"database/sql"

	"github.com/dat1010/go-api/controllers"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers all the API routes
func RegisterRoutes(api *gin.RouterGroup, db *sql.DB) {
	// Pass the database connection to controllers that need it

	// Auth routes
	auth := api.Group("/auth")
	{
		auth.GET("/login", controllers.Login)
		auth.GET("/callback", controllers.Callback)
	}

	// Other routes...
	api.GET("/healthcheck", controllers.GetHealthCheck)
	api.GET("/secrets", controllers.GetSecret)
}
