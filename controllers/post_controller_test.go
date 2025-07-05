package controllers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/dat1010/go-api/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type mockPostService struct {
	CreatePostFunc func(req *models.CreatePostRequest, auth0UserID string) (*models.Post, error)
	GetPostFunc    func(id string) (*models.Post, error)
}

func (m *mockPostService) CreatePost(req *models.CreatePostRequest, auth0UserID string) (*models.Post, error) {
	return m.CreatePostFunc(req, auth0UserID)
}

func (m *mockPostService) GetPost(id string) (*models.Post, error) {
	if m.GetPostFunc != nil {
		return m.GetPostFunc(id)
	}
	return nil, nil
}

func (m *mockPostService) UpdatePost(id string, req *models.UpdatePostRequest, auth0UserID string) (*models.Post, error) {
	return nil, nil
}
func (m *mockPostService) DeletePost(id string, auth0UserID string) error { return nil }
func (m *mockPostService) ListPosts(published *bool, author *string) ([]models.Post, error) {
	return nil, nil
}

func TestCreatePost_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := &mockPostService{
		CreatePostFunc: func(req *models.CreatePostRequest, auth0UserID string) (*models.Post, error) {
			return &models.Post{
				ID:          "1",
				Title:       req.Title,
				Content:     req.Content,
				Auth0UserID: auth0UserID,
				Published:   req.Published,
				Slug:        "test-slug",
			}, nil
		},
	}
	// Set the mock service
	postService = mockService

	r := gin.Default()
	r.POST("/posts", func(c *gin.Context) {
		// Simulate Auth0 user in context
		c.Set("user", validator.RegisteredClaims{Subject: "auth0|testuser"})
		CreatePost(c)
	})

	body := models.CreatePostRequest{
		Title:     "Test Title",
		Content:   "Test Content",
		Published: true,
	}
	jsonBody, _ := json.Marshal(body)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/posts", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp models.Post
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "Test Title", resp.Title)
	assert.Equal(t, "Test Content", resp.Content)
	assert.Equal(t, true, resp.Published)
	assert.Equal(t, "auth0|testuser", resp.Auth0UserID)
}

func TestGetPost_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	expectedPost := &models.Post{
		ID:          "test-id",
		Title:       "Test Post",
		Content:     "Test Content",
		Auth0UserID: "auth0|testuser",
		Published:   true,
		Slug:        "test-post",
	}

	mockService := &mockPostService{
		GetPostFunc: func(id string) (*models.Post, error) {
			return expectedPost, nil
		},
	}

	// Set the mock service
	postService = mockService

	r := gin.Default()
	r.GET("/posts/:id", GetPost)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/posts/test-id", http.NoBody)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp models.Post
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, expectedPost.ID, resp.ID)
	assert.Equal(t, expectedPost.Title, resp.Title)
	assert.Equal(t, expectedPost.Content, resp.Content)
	assert.Equal(t, expectedPost.Auth0UserID, resp.Auth0UserID)
	assert.Equal(t, expectedPost.Published, resp.Published)
	assert.Equal(t, expectedPost.Slug, resp.Slug)
}

func TestGetPost_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &mockPostService{
		GetPostFunc: func(id string) (*models.Post, error) {
			return nil, sql.ErrNoRows
		},
	}

	// Set the mock service
	postService = mockService

	r := gin.Default()
	r.GET("/posts/:id", GetPost)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/posts/nonexistent-id", http.NoBody)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "Post not found", resp["error"])
}

func TestGetPost_InternalServerError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &mockPostService{
		GetPostFunc: func(id string) (*models.Post, error) {
			return nil, errors.New("database connection failed")
		},
	}

	// Set the mock service
	postService = mockService

	r := gin.Default()
	r.GET("/posts/:id", GetPost)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/posts/test-id", http.NoBody)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "database connection failed", resp["error"])
}
