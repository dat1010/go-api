package controllers

import (
	"net/http"

	"github.com/dat1010/go-api/services"
	"github.com/gin-gonic/gin"
)

var userService services.UserService

func SetUserService(s services.UserService) {
	userService = s
}

type CreateUserRequest struct {
	Auth0UserID string `json:"auth0_user_id" binding:"required"`
}

type UpdateUserRoleRequest struct {
	Role string `json:"role" binding:"required"`
}

// @Summary List users
// @Description List users and their roles (superadmin only)
// @Tags admin
// @Produce json
// @Success 200 {array} models.UserWithRole
// @Failure 500 {object} object "Internal server error"
// @Router /admin/users [get]
func ListUsers(c *gin.Context) {
	users, err := userService.ListUsersWithRoles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list users"})
		return
	}
	c.JSON(http.StatusOK, users)
}

// @Summary Create user
// @Description Create a user entry for an Auth0 user id (superadmin only)
// @Tags admin
// @Accept json
// @Produce json
// @Param body body CreateUserRequest true "Create user payload"
// @Success 201 {object} object "User created"
// @Failure 400 {object} object "Bad request"
// @Failure 500 {object} object "Internal server error"
// @Router /admin/users [post]
func CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if err := userService.EnsureUserWithDefaultRole(req.Auth0UserID, "member"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"created": true})
}

// @Summary Update user role
// @Description Set a user's role (superadmin only)
// @Tags admin
// @Accept json
// @Produce json
// @Param id path string true "Auth0 user id"
// @Param body body UpdateUserRoleRequest true "Role payload"
// @Success 200 {object} object "Role updated"
// @Failure 400 {object} object "Bad request"
// @Failure 404 {object} object "Role not found"
// @Failure 500 {object} object "Internal server error"
// @Router /admin/users/{id}/role [patch]
func UpdateUserRole(c *gin.Context) {
	var req UpdateUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	auth0UserID := c.Param("id")
	if auth0UserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing user id"})
		return
	}

	if err := userService.SetUserRole(auth0UserID, req.Role); err != nil {
		if err == services.ErrRoleNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update role"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"updated": true})
}

// @Summary Delete user
// @Description Delete a user entry (superadmin only)
// @Tags admin
// @Produce json
// @Param id path string true "Auth0 user id"
// @Success 200 {object} object "User deleted"
// @Failure 400 {object} object "Bad request"
// @Failure 500 {object} object "Internal server error"
// @Router /admin/users/{id} [delete]
func DeleteUser(c *gin.Context) {
	auth0UserID := c.Param("id")
	if auth0UserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing user id"})
		return
	}
	if err := userService.DeleteUser(auth0UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": true})
}
