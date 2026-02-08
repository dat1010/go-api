package middleware

import (
	"net/http"

	"github.com/dat1010/go-api/repositories"
	"github.com/dat1010/go-api/services"
	"github.com/dat1010/go-api/utils"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func EnsureUserRole(defaultRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		db, ok := c.Get("db")
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "db not available"})
			return
		}
		sqlxDB, ok := db.(*sqlx.DB)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "invalid db"})
			return
		}

		auth0UserID, ok := utils.GetAuth0UserID(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		userRepo := repositories.NewUserRepository(sqlxDB)
		userService := services.NewUserService(userRepo)

		if err := userService.EnsureUserWithDefaultRole(auth0UserID, defaultRole); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to ensure user"})
			return
		}

		c.Next()
	}
}

func RequireRole(roleName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		db, ok := c.Get("db")
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "db not available"})
			return
		}
		sqlxDB, ok := db.(*sqlx.DB)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "invalid db"})
			return
		}

		auth0UserID, ok := utils.GetAuth0UserID(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		userRepo := repositories.NewUserRepository(sqlxDB)
		userService := services.NewUserService(userRepo)

		isAllowed, err := userService.IsUserInRole(auth0UserID, roleName)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to check role"})
			return
		}
		if !isAllowed {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return
		}

		c.Next()
	}
}
