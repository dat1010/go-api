package controllers

import (
	"database/sql"
	"net/http"

	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/dat1010/go-api/models"
	"github.com/dat1010/go-api/services"
	"github.com/dat1010/go-api/utils"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

type CreatePostRequest struct {
	Title     string `json:"title" binding:"required"`
	Content   string `json:"content" binding:"required"`
	Published bool   `json:"published"`
}

type UpdatePostRequest struct {
	Title     string `json:"title"`
	Content   string `json:"content"`
	Published *bool  `json:"published"`
}

var postService services.PostService // This should be set up in main.go or via DI

// SetPostService sets the post service for the controllers
func SetPostService(service services.PostService) {
	postService = service
}

// @Summary Create a new post
// @Description Create a new post with the provided data
// @Tags posts
// @Accept json
// @Produce json
// @Param post body controllers.CreatePostRequest true "Post data"
// @Security Bearer
// @Success 201 {object} models.Post
// @Failure 400 {object} object "Invalid request"
// @Failure 401 {object} object "Unauthorized"
// @Failure 500 {object} object "Internal server error"
// @Router /posts [post]
func CreatePost(c *gin.Context) {
	var req models.CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	auth0UserID, ok := utils.GetAuth0UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	post, err := postService.CreatePost(&req, auth0UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, post)
}

// @Summary Get a post by ID
// @Description Get a post by its ID. This is a public endpoint and does not require authentication.
// @Tags posts
// @Produce json
// @Param id path string true "Post ID"
// @Success 200 {object} models.Post
// @Failure 404 {object} object "Post not found"
// @Failure 500 {object} object "Internal server error"
// @Router /posts/{id} [get]
func GetPost(c *gin.Context) {
	id := c.Param("id")

	post, err := postService.GetPost(id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, post)
}

// @Summary Update a post
// @Description Update an existing post
// @Tags posts
// @Accept json
// @Produce json
// @Param id path string true "Post ID"
// @Param post body controllers.UpdatePostRequest true "Post data"
// @Security Bearer
// @Success 200 {object} models.Post
// @Failure 400 {object} object "Invalid request"
// @Failure 401 {object} object "Unauthorized"
// @Failure 403 {object} object "Forbidden"
// @Failure 404 {object} object "Post not found"
// @Failure 500 {object} object "Internal server error"
// @Router /posts/{id} [put]
func UpdatePost(c *gin.Context) {
	id := c.Param("id")
	var req UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get Auth0 user ID from the JWT claims
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Extract user ID from claims
	registeredClaims, ok := claims.(validator.RegisteredClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid claims format"})
		return
	}

	db := c.MustGet("db").(*sqlx.DB)

	// First, check if the post exists and belongs to the user
	var post models.Post
	err := db.Get(&post, "SELECT * FROM posts WHERE id = ?", id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if post.Auth0UserID != registeredClaims.Subject {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to update this post"})
		return
	}

	// Update the post
	updateQuery := `UPDATE posts SET 
		title = COALESCE(:title, title),
		content = COALESCE(:content, content),
		published = COALESCE(:published, published),
		updated_at = CURRENT_TIMESTAMP
		WHERE id = :id AND auth0_user_id = :auth0_user_id`

	params := map[string]interface{}{
		"id":            id,
		"title":         req.Title,
		"content":       req.Content,
		"published":     req.Published,
		"auth0_user_id": registeredClaims.Subject,
	}

	_, err = db.NamedExec(updateQuery, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get the updated post
	err = db.Get(&post, "SELECT * FROM posts WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, post)
}

// @Summary Delete a post
// @Description Delete a post by its ID
// @Tags posts
// @Produce json
// @Param id path string true "Post ID"
// @Security Bearer
// @Success 204 "No Content"
// @Failure 401 {object} object "Unauthorized"
// @Failure 403 {object} object "Forbidden"
// @Failure 404 {object} object "Post not found"
// @Failure 500 {object} object "Internal server error"
// @Router /posts/{id} [delete]
func DeletePost(c *gin.Context) {
	id := c.Param("id")

	// Get Auth0 user ID from the JWT claims
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Extract user ID from claims
	registeredClaims, ok := claims.(validator.RegisteredClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid claims format"})
		return
	}

	db := c.MustGet("db").(*sqlx.DB)

	// First, check if the post exists and belongs to the user
	var post models.Post
	err := db.Get(&post, "SELECT * FROM posts WHERE id = ?", id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error while checking post: " + err.Error()})
		return
	}

	if post.Auth0UserID != registeredClaims.Subject {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to delete this post"})
		return
	}

	// Delete the post
	result, err := db.Exec("DELETE FROM posts WHERE id = ? AND auth0_user_id = ?", id, registeredClaims.Subject)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error while deleting post: " + err.Error()})
		return
	}

	// Check if any rows were actually deleted
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking rows affected: " + err.Error()})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found or already deleted"})
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary List posts
// @Description Get a list of posts with optional filtering. This is a public endpoint and does not require authentication.
// @Tags posts
// @Produce json
// @Param published query bool false "Filter by published status"
// @Param author query string false "Filter by author ID"
// @Success 200 {array} models.Post
// @Failure 500 {object} object "Internal server error"
// @Router /posts [get]
func ListPosts(c *gin.Context) {
	published := c.Query("published")
	author := c.Query("author")

	db := c.MustGet("db").(*sqlx.DB)
	var posts []models.Post
	var err error

	query := "SELECT * FROM posts WHERE 1=1"
	args := []interface{}{}

	if published != "" {
		query += " AND published = ?"
		args = append(args, published == "true")
	}

	if author != "" {
		query += " AND auth0_user_id = ?"
		args = append(args, author)
	}

	query += " ORDER BY created_at DESC"

	err = db.Select(&posts, query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, posts)
}
