package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/dat1010/go-api/handlers"
	"github.com/dat1010/go-api/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Failed to load the env vars: %v", err)
	}

	url := os.Getenv("TURSO_DATABASE_URL")
	db, err := sql.Open("libsql", url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open db %s: %s", url, err)
		os.Exit(1)
	}
	defer db.Close()

	router := gin.Default()

	// Load HTML templates for Auth0 quick‐start page
	router.LoadHTMLGlob("templates/*.html")

	// Public auth routes
	router.GET("/", handlers.Home)
	router.GET("/login", handlers.Login)
	router.GET("/callback", handlers.Callback)

	// Your existing API routes
	api := router.Group("/api")
	routes.RegisterRoutes(api)

	log.Println("Server listening on http://localhost:3000/")
	router.Run(":3000")
}
