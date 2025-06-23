package controllers

import (
	"bytes"
	"encoding/json"
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
}

func (m *mockPostService) CreatePost(req *models.CreatePostRequest, auth0UserID string) (*models.Post, error) {
	return m.CreatePostFunc(req, auth0UserID)
}
func (m *mockPostService) GetPost(id string) (*models.Post, error) { return nil, nil }
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
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Test Title", resp.Title)
	assert.Equal(t, "Test Content", resp.Content)
	assert.Equal(t, true, resp.Published)
	assert.Equal(t, "auth0|testuser", resp.Auth0UserID)
}
