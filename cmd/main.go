package main

import (
	"os"

	_ "github.com/dat1010/go-api/docs"
	"github.com/dat1010/go-api/routes"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	router := gin.Default()

	api := router.Group("/api")
	routes.RegisterRoutes(api)

	bindAddr := os.Getenv("BIND_ADDR")
	if bindAddr == "" {
		bindAddr = "0.0.0.0:8080"
	}

	// serve Swagger UI
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.Run(bindAddr)
}
