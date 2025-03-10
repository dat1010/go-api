package main

import (
	"github.com/dat1010/go-api/routes"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	api := router.Group("/api")
	routes.RegisterRoutes(api)

	router.Run("0.0.0.0:8080")
}
