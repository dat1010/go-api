package routes

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/healthcheck", controllers.GetHealthCheck)
}
