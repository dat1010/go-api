package models

import "time"

type User struct {
	Auth0UserID string    `json:"auth0_user_id" db:"auth0_user_id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type UserWithRole struct {
	Auth0UserID string `json:"auth0_user_id" db:"auth0_user_id"`
	Role        string `json:"role" db:"role"`
}
