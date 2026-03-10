package services

import (
	"database/sql"
	"testing"
	"time"

	"github.com/dat1010/go-api/models"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
)

type mockMoodRepository struct {
	ListActiveTagsFunc     func() ([]models.MoodTag, error)
	CreateTagFunc          func(tag *models.MoodTag) error
	GetActiveTagsByIDsFunc func(ids []string) ([]models.MoodTag, error)
	CreateEntryFunc        func(entry *models.MoodEntry, tagIDs []string) error
	ListEntriesFunc        func(params models.ListMoodEntriesParams) ([]models.MoodEntry, error)
	GetEntryByIDFunc       func(id string) (*models.MoodEntry, error)
	UpdateEntryFunc        func(id string, note *string, tagIDs []string) error
}

func (m *mockMoodRepository) ListActiveTags() ([]models.MoodTag, error) {
	return m.ListActiveTagsFunc()
}

func (m *mockMoodRepository) CreateTag(tag *models.MoodTag) error {
	return m.CreateTagFunc(tag)
}

func (m *mockMoodRepository) GetActiveTagsByIDs(ids []string) ([]models.MoodTag, error) {
	return m.GetActiveTagsByIDsFunc(ids)
}

func (m *mockMoodRepository) CreateEntry(entry *models.MoodEntry, tagIDs []string) error {
	return m.CreateEntryFunc(entry, tagIDs)
}

func (m *mockMoodRepository) ListEntries(params models.ListMoodEntriesParams) ([]models.MoodEntry, error) {
	return m.ListEntriesFunc(params)
}

func (m *mockMoodRepository) GetEntryByID(id string) (*models.MoodEntry, error) {
	return m.GetEntryByIDFunc(id)
}

func (m *mockMoodRepository) UpdateEntry(id string, note *string, tagIDs []string) error {
	return m.UpdateEntryFunc(id, note, tagIDs)
}

func TestCreateMoodTagNormalizesName(t *testing.T) {
	repo := &mockMoodRepository{
		CreateTagFunc: func(tag *models.MoodTag) error {
			assert.Equal(t, "hopeful", tag.Name)
			assert.True(t, tag.IsActive)
			return nil
		},
	}

	service := NewMoodService(repo)
	tag, err := service.CreateMoodTag(&models.CreateMoodTagRequest{Name: "  Hopeful  "})

	assert.NoError(t, err)
	if err != nil {
		return
	}
	assert.Equal(t, "hopeful", tag.Name)
}

func TestCreateMoodTagDuplicateFails(t *testing.T) {
	repo := &mockMoodRepository{
		CreateTagFunc: func(tag *models.MoodTag) error {
			return &pgconn.PgError{Code: "23505"}
		},
	}

	service := NewMoodService(repo)
	_, err := service.CreateMoodTag(&models.CreateMoodTagRequest{Name: "hopeful"})

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrMoodTagExists)
}

func TestCreateMoodEntryDedupesTagIDsAndTrimsNote(t *testing.T) {
	repo := &mockMoodRepository{
		GetActiveTagsByIDsFunc: func(ids []string) ([]models.MoodTag, error) {
			assert.Equal(t, []string{
				"0b35b2d4-4d68-4c17-96bc-1c6e1f4dbf9a",
				"6c39156f-3af7-4ca0-aef1-7b9f06f2cf29",
			}, ids)
			return []models.MoodTag{
				{ID: ids[0], Name: "calm", IsActive: true},
				{ID: ids[1], Name: "tired", IsActive: true},
			}, nil
		},
		CreateEntryFunc: func(entry *models.MoodEntry, tagIDs []string) error {
			assert.Len(t, tagIDs, 2)
			assert.NotNil(t, entry.Note)
			if entry.Note == nil {
				return nil
			}
			assert.Equal(t, "Felt better later", *entry.Note)
			return nil
		},
		GetEntryByIDFunc: func(id string) (*models.MoodEntry, error) {
			note := "Felt better later"
			return &models.MoodEntry{
				ID:        id,
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
				Note:      &note,
				Tags: []models.MoodTag{
					{ID: "0b35b2d4-4d68-4c17-96bc-1c6e1f4dbf9a", Name: "calm"},
					{ID: "6c39156f-3af7-4ca0-aef1-7b9f06f2cf29", Name: "tired"},
				},
			}, nil
		},
	}

	service := NewMoodService(repo)
	note := "  Felt better later  "
	entry, err := service.CreateMoodEntry(&models.CreateMoodEntryRequest{
		TagIDs: []string{
			"0b35b2d4-4d68-4c17-96bc-1c6e1f4dbf9a",
			"0b35b2d4-4d68-4c17-96bc-1c6e1f4dbf9a",
			"6c39156f-3af7-4ca0-aef1-7b9f06f2cf29",
		},
		Note: &note,
	})

	assert.NoError(t, err)
	if err != nil {
		return
	}
	assert.Len(t, entry.Tags, 2)
}

func TestCreateMoodEntryRequiresAtLeastOneTag(t *testing.T) {
	service := NewMoodService(&mockMoodRepository{})
	_, err := service.CreateMoodEntry(&models.CreateMoodEntryRequest{})

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrMoodTagsRequired)
}

func TestUpdateMoodEntryNotFound(t *testing.T) {
	repo := &mockMoodRepository{
		GetActiveTagsByIDsFunc: func(ids []string) ([]models.MoodTag, error) {
			return []models.MoodTag{{ID: ids[0], Name: "calm", IsActive: true}}, nil
		},
		UpdateEntryFunc: func(id string, note *string, tagIDs []string) error {
			return sql.ErrNoRows
		},
	}

	service := NewMoodService(repo)
	_, err := service.UpdateMoodEntry("0b35b2d4-4d68-4c17-96bc-1c6e1f4dbf9a", &models.UpdateMoodEntryRequest{
		TagIDs: []string{"0b35b2d4-4d68-4c17-96bc-1c6e1f4dbf9a"},
	})

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrMoodEntryNotFound)
}
