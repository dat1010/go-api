package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/gin-gonic/gin"
)

// CustomClaims contains custom data we want from the token.
type CustomClaims struct {
	Scope string `json:"scope"`
}

// Validate does nothing for this example, but we need
// it to satisfy validator.CustomClaims interface.
func (c CustomClaims) Validate(ctx context.Context) error {
	return nil
}

func Auth0() gin.HandlerFunc {
	domain := os.Getenv("AUTH0_DOMAIN")
	audience := os.Getenv("AUTH0_AUDIENCE")
	
	if domain == "" {
		panic("AUTH0_DOMAIN environment variable not set")
	}
	if audience == "" {
		panic("AUTH0_AUDIENCE environment variable not set")
	}

	log.Printf("Setting up Auth0 middleware with domain: %s and audience: %s", domain, audience)

	// Set up the validator
	jwtValidator, err := validator.New(
		func(ctx context.Context) (interface{}, error) {
			return nil, nil // For RS256, we don't need a key provider
		},
		validator.RS256,
		"https://"+domain+"/",
		[]string{audience},
		validator.WithCustomClaims(func() validator.CustomClaims {
			return &CustomClaims{}
		}),
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to set up the validator: %v", err))
	}

	middleware := jwtmiddleware.New(jwtValidator.ValidateToken)

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format must be Bearer {token}"})
			return
		}

		token := parts[1]
		log.Printf("Received token: %s", token)

		// Create a new request with the token
		req := c.Request.Clone(c.Request.Context())
		req.Header.Set("Authorization", authHeader)

		// Validate the token
		var validationError error
		var handler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
			c.Request = r
			validationError = nil
		}

		middleware.CheckJWT(handler).ServeHTTP(c.Writer, req)

		if validationError != nil {
			log.Printf("Token validation error: %v", validationError)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
				"details": validationError.Error(),
			})
			return
		}

		// Get the claims from the token
		claims, ok := c.Request.Context().Value(jwtmiddleware.ContextKey{}).(*validator.ValidatedClaims)
		if !ok {
			log.Printf("Failed to extract claims from token")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token format",
				"details": "Could not extract claims from token",
			})
			return
		}

		log.Printf("Token validated successfully. Claims: %+v", claims)
		c.Set("user", claims.RegisteredClaims)
		c.Next()
	}
}
