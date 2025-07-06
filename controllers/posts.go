package controllers

import (
	"database/sql"
	"net/http"

	"github.com/dat1010/go-api/models"
	"github.com/dat1010/go-api/services"
	"github.com/dat1010/go-api/utils"
	"github.com/gin-gonic/gin"
)

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
// @Param post body models.UpdatePostRequest true "Post data"
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
	var req models.UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	auth0UserID, ok := utils.GetAuth0UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	post, err := postService.UpdatePost(id, &req, auth0UserID)
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

	auth0UserID, ok := utils.GetAuth0UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	err := postService.DeletePost(id, auth0UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary List posts
// @Description Get a list of posts with optional filtering. This is a public endpoint and does not require authentication.
// @Tags posts
// @Produce json
// @Param author query string false "Filter by author ID"
// @Success 200 {array} models.Post
// @Failure 500 {object} object "Internal server error"
// @Router /posts [get]
func ListPosts(c *gin.Context) {
	author := c.Query("author")

	var authorFilter *string
	if author != "" {
		authorFilter = &author
	}

	posts, err := postService.ListPosts(authorFilter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, posts)
}
