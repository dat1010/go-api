package middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/jwks"
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

// Add this type at the top of the file, after the imports
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
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

	// Set up the key provider
	issuerURL, err := url.Parse("https://" + domain + "/")
	if err != nil {
		panic(fmt.Sprintf("Failed to parse issuer URL: %v", err))
	}

	provider := jwks.NewCachingProvider(
		issuerURL,
		5*time.Minute,
	)

	// Set up the validator
	jwtValidator, err := validator.New(
		provider.KeyFunc,
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

		// Create a new request with the token
		req := c.Request.Clone(c.Request.Context())
		req.Header.Set("Authorization", authHeader)

		// Create a response writer that can capture validation failures
		recorder := &responseRecorder{ResponseWriter: c.Writer}
		var handler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
			c.Request = r
		}

		middleware.CheckJWT(handler).ServeHTTP(recorder, req)

		if recorder.statusCode >= 400 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
				"details": "Token validation failed",
			})
			return
		}

		// Get the claims from the token
		claims, ok := c.Request.Context().Value(jwtmiddleware.ContextKey{}).(*validator.ValidatedClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token format",
				"details": "Could not extract claims from token",
			})
			return
		}

		c.Set("user", claims.RegisteredClaims)
		c.Next()
	}
}
