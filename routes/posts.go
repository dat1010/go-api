package routes

import (
	"github.com/dat1010/go-api/controllers"
	"github.com/dat1010/go-api/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterPostRoutes(r *gin.RouterGroup) {
	posts := r.Group("/posts")
	{
		// Public routes
		posts.GET("", controllers.ListPosts)
		posts.GET("/:id", controllers.GetPost)

		// Protected routes
		posts.Use(middleware.Auth0())
		posts.Use(middleware.EnsureUserRole("member"))
		posts.POST("", controllers.CreatePost)
		posts.PUT("/:id", controllers.UpdatePost)
		posts.DELETE("/:id", controllers.DeletePost)
	}
}
