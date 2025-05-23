package middleware

import (
	"context"
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
	if domain == "" {
		panic("AUTH0_DOMAIN environment variable not set")
	}

	// Set up the validator
	jwtValidator, err := validator.New(
		func(ctx context.Context) (interface{}, error) {
			return []byte(os.Getenv("AUTH0_CLIENT_SECRET")), nil
		},
		validator.HS256,
		"https://"+domain+"/",
		[]string{os.Getenv("AUTH0_AUDIENCE")},
	)
	if err != nil {
		panic(err)
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

		// Validate the token
		encounteredError := false
		var handler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
			c.Request = r
		}

		middleware.CheckJWT(handler).ServeHTTP(c.Writer, req)

		if encounteredError {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		// Get the claims from the token
		token := c.Request.Context().Value(jwtmiddleware.ContextKey{}).(*validator.ValidatedClaims)
		c.Set("user", token.RegisteredClaims)
		c.Next()
	}
}
