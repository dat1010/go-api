package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/dat1010/go-api/docs"
	"github.com/dat1010/go-api/routes"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Initialize Turso database connection
	dbURL := os.Getenv("TURSO_DATABASE_URL")
	authToken := os.Getenv("TURSO_AUTH_TOKEN")
	if dbURL == "" || authToken == "" {
		log.Fatal("TURSO_DATABASE_URL and TURSO_AUTH_TOKEN environment variables must be set")
	}

	connectionURL := fmt.Sprintf("%s?authToken=%s", dbURL, authToken)
	
	// Configure GORM logger
	gormLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	// Initialize GORM with SQLite driver
	db, err := gorm.Open(sqlite.Open(connectionURL), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Test the connection
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get database instance: %v", err)
	}
	defer sqlDB.Close()

	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Successfully connected to Turso database with GORM")

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

	// Make db available to routes
	api := router.Group("/api")
	routes.RegisterRoutes(api, db)

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

	// Run HTTP server
	log.Printf("Starting HTTP server on %s", bindAddr)
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-ctx.Done()
	log.Println("Shutting down server...")

	// Create a timeout context for shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown the server
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server forced to shutdown: %v", err)
	}

	log.Println("Server gracefully stopped")
}
