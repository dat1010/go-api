package routes

import (
    "github.com/dat1010/go-api/controllers"
    "github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup) {
    r.GET("/healthcheck", controllers.GetHealthCheck)
}