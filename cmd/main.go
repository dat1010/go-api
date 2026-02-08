//go:generate swag init --generalInfo main.go --output ../docs
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
	"github.com/dat1010/go-api/controllers"
	_ "github.com/dat1010/go-api/docs"
	"github.com/dat1010/go-api/middleware"
	"github.com/dat1010/go-api/repositories"
	"github.com/dat1010/go-api/routes"
	"github.com/dat1010/go-api/services"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           Go API
// @version         1.0
// @description     A simple API written in Golang with AWS EventBridge integration
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      api.nofeed.zone
// @BasePath  /api

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	if err := config.RunMigrations(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize database connection
	db, err := config.NewDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize repositories and services
	postRepo := repositories.NewPostRepository(db)
	postService := services.NewPostService(postRepo)
	userRepo := repositories.NewUserRepository(db)
	userService := services.NewUserService(userRepo)

	// Set the post service for controllers
	controllers.SetPostService(postService)
	controllers.SetUserService(userService)

	router := gin.Default()

	// Add CORS middleware
	router.Use(middleware.CORS())

	// Trust all proxies (for older Gin versions)
	if err := router.SetTrustedProxies(nil); err != nil {
		log.Printf("Warning: failed to set trusted proxies: %v", err)
	}
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

	// serve Swagger UI with custom configuration
	swaggerHandler := ginSwagger.DisablingWrapHandler(swaggerFiles.Handler, "DISABLE_SWAGGER_HTTP_HANDLER")
	router.GET("/swagger/*any", swaggerHandler)

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
