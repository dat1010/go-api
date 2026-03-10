package routes

import (
	"github.com/dat1010/go-api/controllers"
	"github.com/gin-gonic/gin"
)

func RegisterMoodRoutes(r *gin.RouterGroup) {
	moods := r.Group("")

	moods.GET("/mood-tags", controllers.GetMoodTags)
	moods.POST("/mood-tags", controllers.CreateMoodTag)
	moods.GET("/mood-entries", controllers.GetMoodEntries)
	moods.GET("/mood-entries/:id", controllers.GetMoodEntry)
	moods.POST("/mood-entries", controllers.CreateMoodEntry)
	moods.PUT("/mood-entries/:id", controllers.UpdateMoodEntry)
}
