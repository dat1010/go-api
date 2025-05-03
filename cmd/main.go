package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dat1010/go-api/config"
	_ "github.com/dat1010/go-api/docs"
	"github.com/dat1010/go-api/routes"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	// Initialize database connection
	db, err := config.NewDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	router := gin.Default()

	// Trust all proxies (for older Gin versions)
	router.SetTrustedProxies(nil)
	router.ForwardedByClientIP = true
	router.AppEngine = false

	// Add middleware to force HTTPS in URLs when behind load balancer
	router.Use(func(c *gin.Context) {
		if c.GetHeader("X-Forwarded-Proto") == "https" {
			c.Request.URL.Scheme = "https"
		}
		c.Next()
	})

	// Make db available to all routes
	router.Use(func(c *gin.Context) {
		c.Set("db", db)
		c.Next()
	})

	// Register routes
	api := router.Group("/api")
	routes.RegisterRoutes(api)
	routes.RegisterPostRoutes(api)

	// serve Swagger UI
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Get the HTTP bind address
	bindAddr := os.Getenv("BIND_ADDR")
	if bindAddr == "" {
		bindAddr = "0.0.0.0:8080"
	}

	// Create a context for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Create an HTTP server
	httpServer := &http.Server{
		Addr:    bindAddr,
		Handler: router,
	}

	// Start the server in a goroutine
	go func() {
		log.Printf("Server starting on %s", bindAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-ctx.Done()

	// Create a shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}
