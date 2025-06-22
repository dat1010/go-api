package utils

import (
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/gin-gonic/gin"
)

// GetAuth0UserID extracts the Auth0 user ID from Gin context
func GetAuth0UserID(c *gin.Context) (string, bool) {
	claims, exists := c.Get("user")
	if !exists {
		return "", false
	}
	registeredClaims, ok := claims.(validator.RegisteredClaims)
	if !ok {
		return "", false
	}
	return registeredClaims.Subject, true
}
