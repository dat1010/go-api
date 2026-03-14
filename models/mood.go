package models

import "time"

type MoodTag struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	IsActive  bool      `json:"isActive" db:"is_active"`
	CreatedAt time.Time `json:"createdAt,omitempty" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt,omitempty" db:"updated_at"`
}

type MoodEntry struct {
	ID          string    `json:"id" db:"id"`
	Auth0UserID string    `json:"-" db:"auth0_user_id"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time `json:"updatedAt" db:"updated_at"`
	Note        *string   `json:"note" db:"note"`
	Tags        []MoodTag `json:"tags"`
}

type CreateMoodTagRequest struct {
	Name string `json:"name"`
}

type CreateMoodEntryRequest struct {
	TagIDs []string `json:"tagIds"`
	Note   *string  `json:"note"`
}

type UpdateMoodEntryRequest struct {
	TagIDs []string `json:"tagIds"`
	Note   *string  `json:"note"`
}

type ListMoodEntriesParams struct {
	Limit int
	From  *time.Time
	To    *time.Time
}
