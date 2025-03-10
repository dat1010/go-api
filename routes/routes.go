package routes

import (
	"github.com/dat1010/go-api/controllers"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/healthcheck", controllers.GetHealthCheck)
	router.GET("/stats/weekly", controllers.GetWeeklyStats)
}
