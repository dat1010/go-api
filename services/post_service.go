package services

import (
	"database/sql"
	"time"

	"github.com/dat1010/go-api/models"
	"github.com/dat1010/go-api/repositories"
	"github.com/google/uuid"
)

type PostService interface {
	CreatePost(req *models.CreatePostRequest, auth0UserID string) (*models.Post, error)
	GetPost(id string) (*models.Post, error)
	UpdatePost(id string, req *models.UpdatePostRequest, auth0UserID string) (*models.Post, error)
	DeletePost(id, auth0UserID string) error
	ListPosts(published *bool, author *string) ([]models.Post, error)
}

type postService struct {
	postRepo repositories.PostRepository
}

func NewPostService(postRepo repositories.PostRepository) PostService {
	return &postService{postRepo: postRepo}
}

func (s *postService) CreatePost(req *models.CreatePostRequest, auth0UserID string) (*models.Post, error) {
	post := &models.Post{
		ID:          uuid.New().String(),
		Title:       req.Title,
		Content:     req.Content,
		Auth0UserID: auth0UserID,
		Published:   req.Published,
		Slug:        generateSlug(req.Title),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := s.postRepo.Create(post)
	if err != nil {
		return nil, err
	}

	return post, nil
}

func (s *postService) GetPost(id string) (*models.Post, error) {
	post, err := s.postRepo.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}
	return post, nil
}

func (s *postService) UpdatePost(id string, req *models.UpdatePostRequest, auth0UserID string) (*models.Post, error) {
	// First, check if the post exists and belongs to the user
	existingPost, err := s.postRepo.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}

	if existingPost.Auth0UserID != auth0UserID {
		return nil, sql.ErrNoRows // Use this to indicate permission denied
	}

	// Prepare updates
	updates := map[string]interface{}{
		"auth0_user_id": auth0UserID,
	}

	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Content != "" {
		updates["content"] = req.Content
	}
	if req.Published != nil {
		updates["published"] = *req.Published
	}

	// Update the post
	err = s.postRepo.Update(id, updates)
	if err != nil {
		return nil, err
	}

	// Get the updated post
	return s.postRepo.GetByID(id)
}

func (s *postService) DeletePost(id, auth0UserID string) error {
	// First, check if the post exists and belongs to the user
	existingPost, err := s.postRepo.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return err
		}
		return err
	}

	if existingPost.Auth0UserID != auth0UserID {
		return sql.ErrNoRows // Use this to indicate permission denied
	}

	return s.postRepo.Delete(id, auth0UserID)
}

func (s *postService) ListPosts(published *bool, author *string) ([]models.Post, error) {
	switch {
	case published != nil && author != nil:
		return s.postRepo.ListByAuthorAndPublished(*author, *published)
	case published != nil:
		return s.postRepo.ListByPublished(*published)
	case author != nil:
		return s.postRepo.ListByAuthor(*author)
	default:
		return s.postRepo.List()
	}
}

// Helper function to generate a URL-friendly slug from a title.
func generateSlug(title string) string {
	// TODO: Implement proper slug generation.
	// For now, just return a simple slug.
	return title
}
