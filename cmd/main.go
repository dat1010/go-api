package main

import (
	"os"

	"github.com/dat1010/go-api/routes"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	api := router.Group("/api")
	routes.RegisterRoutes(api)

	// pick bind address from BIND_ADDR (default to 0.0.0.0:8080)
	bindAddr := os.Getenv("BIND_ADDR")
	if bindAddr == "" {
		bindAddr = "0.0.0.0:8080"
	}
	router.Run(bindAddr)
}
