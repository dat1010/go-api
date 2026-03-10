package services

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/dat1010/go-api/models"
	"github.com/dat1010/go-api/repositories"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	defaultMoodEntriesLimit = 30
	maxMoodEntriesLimit     = 100
	maxMoodTagNameLength    = 50
	maxMoodEntryNoteLength  = 2000
)

var (
	ErrMoodTagExists       = errors.New("mood tag already exists")
	ErrMoodEntryNotFound   = errors.New("mood entry not found")
	ErrInvalidMoodTagIDs   = errors.New("all tag IDs must exist and be active")
	ErrMoodTagsRequired    = errors.New("at least one tag is required")
	ErrMoodTagNameRequired = errors.New("tag name is required")
)

type MoodService interface {
	ListMoodTags() ([]models.MoodTag, error)
	CreateMoodTag(req *models.CreateMoodTagRequest) (*models.MoodTag, error)
	ListMoodEntries(params models.ListMoodEntriesParams) ([]models.MoodEntry, error)
	GetMoodEntry(id string) (*models.MoodEntry, error)
	CreateMoodEntry(req *models.CreateMoodEntryRequest) (*models.MoodEntry, error)
	UpdateMoodEntry(id string, req *models.UpdateMoodEntryRequest) (*models.MoodEntry, error)
}

type moodService struct {
	repo repositories.MoodRepository
}

func NewMoodService(repo repositories.MoodRepository) MoodService {
	return &moodService{repo: repo}
}

func (s *moodService) ListMoodTags() ([]models.MoodTag, error) {
	return s.repo.ListActiveTags()
}

func (s *moodService) CreateMoodTag(req *models.CreateMoodTagRequest) (*models.MoodTag, error) {
	name, err := normalizeMoodTagName(req.Name)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	tag := &models.MoodTag{
		ID:        uuid.New().String(),
		Name:      name,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.CreateTag(tag); err != nil {
		if isUniqueViolation(err) {
			return nil, ErrMoodTagExists
		}
		return nil, err
	}

	return tag, nil
}

func (s *moodService) ListMoodEntries(params models.ListMoodEntriesParams) ([]models.MoodEntry, error) {
	if params.Limit <= 0 {
		params.Limit = defaultMoodEntriesLimit
	}
	if params.Limit > maxMoodEntriesLimit {
		params.Limit = maxMoodEntriesLimit
	}

	return s.repo.ListEntries(params)
}

func (s *moodService) GetMoodEntry(id string) (*models.MoodEntry, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, err
	}

	entry, err := s.repo.GetEntryByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrMoodEntryNotFound
		}
		return nil, err
	}

	return entry, nil
}

func (s *moodService) CreateMoodEntry(req *models.CreateMoodEntryRequest) (*models.MoodEntry, error) {
	tagIDs, err := s.validateMoodEntryInput(req.TagIDs, req.Note)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	entry := &models.MoodEntry{
		ID:        uuid.New().String(),
		CreatedAt: now,
		UpdatedAt: now,
		Note:      normalizeMoodNote(req.Note),
	}

	if err := s.repo.CreateEntry(entry, tagIDs); err != nil {
		return nil, err
	}

	return s.repo.GetEntryByID(entry.ID)
}

func (s *moodService) UpdateMoodEntry(id string, req *models.UpdateMoodEntryRequest) (*models.MoodEntry, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, err
	}

	tagIDs, err := s.validateMoodEntryInput(req.TagIDs, req.Note)
	if err != nil {
		return nil, err
	}

	if err := s.repo.UpdateEntry(id, normalizeMoodNote(req.Note), tagIDs); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrMoodEntryNotFound
		}
		return nil, err
	}

	return s.repo.GetEntryByID(id)
}

func (s *moodService) validateMoodEntryInput(tagIDs []string, note *string) ([]string, error) {
	if len(tagIDs) == 0 {
		return nil, ErrMoodTagsRequired
	}

	dedupedTagIDs := dedupeMoodTagIDs(tagIDs)
	if len(dedupedTagIDs) == 0 {
		return nil, ErrMoodTagsRequired
	}

	for _, id := range dedupedTagIDs {
		if _, err := uuid.Parse(id); err != nil {
			return nil, err
		}
	}

	normalizedNote := normalizeMoodNote(note)
	if normalizedNote != nil && len(*normalizedNote) > maxMoodEntryNoteLength {
		return nil, errors.New("note must be 2000 characters or fewer")
	}

	tags, err := s.repo.GetActiveTagsByIDs(dedupedTagIDs)
	if err != nil {
		return nil, err
	}
	if len(tags) != len(dedupedTagIDs) {
		return nil, ErrInvalidMoodTagIDs
	}

	return dedupedTagIDs, nil
}

func normalizeMoodTagName(name string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(name))
	if normalized == "" {
		return "", ErrMoodTagNameRequired
	}
	if len(normalized) > maxMoodTagNameLength {
		return "", errors.New("tag name must be 50 characters or fewer")
	}
	return normalized, nil
}

func normalizeMoodNote(note *string) *string {
	if note == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*note)
	return &trimmed
}

func dedupeMoodTagIDs(tagIDs []string) []string {
	seen := make(map[string]struct{}, len(tagIDs))
	deduped := make([]string, 0, len(tagIDs))

	for _, tagID := range tagIDs {
		trimmed := strings.TrimSpace(tagID)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		deduped = append(deduped, trimmed)
	}

	return deduped
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
