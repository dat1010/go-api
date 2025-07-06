package repositories

import (
	"github.com/dat1010/go-api/models"
	"github.com/jmoiron/sqlx"
)

type PostRepository interface {
	Create(post *models.Post) error
	GetByID(id string) (*models.Post, error)
	Update(id string, updates map[string]interface{}) error
	Delete(id, auth0UserID string) error
	List() ([]models.Post, error)
	ListByAuthor(auth0UserID string) ([]models.Post, error)
}

type postRepository struct {
	db *sqlx.DB
}

func NewPostRepository(db *sqlx.DB) PostRepository {
	return &postRepository{db: db}
}

func (r *postRepository) Create(post *models.Post) error {
	query := `INSERT INTO posts (id, title, content, auth0_user_id, created_at, updated_at, slug)
			  VALUES (:id, :title, :content, :auth0_user_id, :created_at, :updated_at, :slug)`

	_, err := r.db.NamedExec(query, post)
	return err
}

func (r *postRepository) GetByID(id string) (*models.Post, error) {
	var post models.Post
	err := r.db.Get(&post, "SELECT * FROM posts WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *postRepository) Update(id string, updates map[string]interface{}) error {
	query := `UPDATE posts SET 
		title = COALESCE(:title, title),
		content = COALESCE(:content, content),
		updated_at = CURRENT_TIMESTAMP
		WHERE id = :id AND auth0_user_id = :auth0_user_id`

	updates["id"] = id
	_, err := r.db.NamedExec(query, updates)
	return err
}

func (r *postRepository) Delete(id, auth0UserID string) error {
	_, err := r.db.Exec("DELETE FROM posts WHERE id = ? AND auth0_user_id = ?", id, auth0UserID)
	return err
}

func (r *postRepository) List() ([]models.Post, error) {
	var posts []models.Post
	err := r.db.Select(&posts, "SELECT * FROM posts ORDER BY created_at DESC")
	return posts, err
}

func (r *postRepository) ListByAuthor(auth0UserID string) ([]models.Post, error) {
	var posts []models.Post
	err := r.db.Select(&posts, "SELECT * FROM posts WHERE auth0_user_id = ? ORDER BY created_at DESC", auth0UserID)
	return posts, err
}
