package repositories

import (
	"database/sql"

	"github.com/dat1010/go-api/models"
	"github.com/jmoiron/sqlx"
)

type UserRepository interface {
	EnsureUser(auth0UserID string) error
	GetUserRole(auth0UserID string) (string, error)
	SetUserRole(auth0UserID, roleName string) error
	ListUsersWithRoles() ([]models.UserWithRole, error)
	DeleteUser(auth0UserID string) error
	IsUserInRole(auth0UserID, roleName string) (bool, error)
}

type userRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) EnsureUser(auth0UserID string) error {
	_, err := r.db.Exec(
		`INSERT INTO users (auth0_user_id) VALUES ($1)
		 ON CONFLICT (auth0_user_id) DO NOTHING`,
		auth0UserID,
	)
	return err
}

func (r *userRepository) GetUserRole(auth0UserID string) (string, error) {
	var role string
	err := r.db.Get(&role, `
		SELECT r.name
		FROM user_roles ur
		JOIN roles r ON r.id = ur.role_id
		WHERE ur.auth0_user_id = $1
	`, auth0UserID)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return role, err
}

func (r *userRepository) SetUserRole(auth0UserID, roleName string) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var roleID int
	if err = tx.Get(&roleID, `SELECT id FROM roles WHERE name = $1`, roleName); err != nil {
		return err
	}

	if _, err = tx.Exec(
		`INSERT INTO user_roles (auth0_user_id, role_id)
		 VALUES ($1, $2)
		 ON CONFLICT (auth0_user_id) DO UPDATE SET role_id = EXCLUDED.role_id, assigned_at = NOW()`,
		auth0UserID, roleID,
	); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *userRepository) ListUsersWithRoles() ([]models.UserWithRole, error) {
	var users []models.UserWithRole
	err := r.db.Select(&users, `
		SELECT u.auth0_user_id,
			   COALESCE(r.name, '') AS role
		FROM users u
		LEFT JOIN user_roles ur ON ur.auth0_user_id = u.auth0_user_id
		LEFT JOIN roles r ON r.id = ur.role_id
		ORDER BY u.created_at DESC
	`)
	return users, err
}

func (r *userRepository) DeleteUser(auth0UserID string) error {
	_, err := r.db.Exec(`DELETE FROM users WHERE auth0_user_id = $1`, auth0UserID)
	return err
}

func (r *userRepository) IsUserInRole(auth0UserID, roleName string) (bool, error) {
	var exists bool
	err := r.db.Get(&exists, `
		SELECT EXISTS(
			SELECT 1
			FROM user_roles ur
			JOIN roles r ON r.id = ur.role_id
			WHERE ur.auth0_user_id = $1 AND r.name = $2
		)
	`, auth0UserID, roleName)
	return exists, err
}
