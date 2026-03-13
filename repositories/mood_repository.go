package repositories

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/dat1010/go-api/models"
	"github.com/jmoiron/sqlx"
)

type MoodRepository interface {
	ListActiveTags() ([]models.MoodTag, error)
	CreateTag(tag *models.MoodTag) error
	GetActiveTagsByIDs(ids []string) ([]models.MoodTag, error)
	CreateEntry(entry *models.MoodEntry, tagIDs []string) error
	ListEntries(params models.ListMoodEntriesParams) ([]models.MoodEntry, error)
	ListEntriesForAnalytics(start, end time.Time) ([]models.MoodEntry, error)
	GetEntryByID(id string) (*models.MoodEntry, error)
	UpdateEntry(id string, note *string, tagIDs []string) error
}

type moodRepository struct {
	db *sqlx.DB
}

func NewMoodRepository(db *sqlx.DB) MoodRepository {
	return &moodRepository{db: db}
}

func (r *moodRepository) ListActiveTags() ([]models.MoodTag, error) {
	var tags []models.MoodTag
	err := r.db.Select(&tags, `
		SELECT id, name, is_active, created_at, updated_at
		FROM mood_tags
		WHERE is_active = TRUE
		ORDER BY name ASC
	`)
	return tags, err
}

func (r *moodRepository) CreateTag(tag *models.MoodTag) error {
	_, err := r.db.NamedExec(`
		INSERT INTO mood_tags (id, name, is_active, created_at, updated_at)
		VALUES (:id, :name, :is_active, :created_at, :updated_at)
	`, tag)
	return err
}

func (r *moodRepository) GetActiveTagsByIDs(ids []string) ([]models.MoodTag, error) {
	if len(ids) == 0 {
		return []models.MoodTag{}, nil
	}

	query, args, err := sqlx.In(`
		SELECT id, name, is_active, created_at, updated_at
		FROM mood_tags
		WHERE is_active = TRUE AND id IN (?)
	`, ids)
	if err != nil {
		return nil, err
	}

	query = r.db.Rebind(query)
	var tags []models.MoodTag
	if err := r.db.Select(&tags, query, args...); err != nil {
		return nil, err
	}
	return tags, nil
}

func (r *moodRepository) CreateEntry(entry *models.MoodEntry, tagIDs []string) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.Exec(`
		INSERT INTO mood_entries (id, created_at, updated_at, note)
		VALUES ($1, $2, $3, $4)
	`, entry.ID, entry.CreatedAt, entry.UpdatedAt, entry.Note); err != nil {
		return err
	}

	if err = insertMoodEntryTags(tx, entry.ID, tagIDs); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *moodRepository) ListEntries(params models.ListMoodEntriesParams) ([]models.MoodEntry, error) {
	baseQuery := `
		SELECT id, created_at, updated_at, note
		FROM mood_entries
	`

	var conditions []string
	args := []interface{}{}

	if params.From != nil {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", len(args)+1))
		args = append(args, *params.From)
	}
	if params.To != nil {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", len(args)+1))
		args = append(args, *params.To)
	}
	if len(conditions) > 0 {
		baseQuery += " WHERE " + strings.Join(conditions, " AND ")
	}

	baseQuery += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d", len(args)+1)
	args = append(args, params.Limit)

	var entries []models.MoodEntry
	if err := r.db.Select(&entries, baseQuery, args...); err != nil {
		return nil, err
	}

	return r.hydrateEntries(entries)
}

func (r *moodRepository) ListEntriesForAnalytics(start, end time.Time) ([]models.MoodEntry, error) {
	var entries []models.MoodEntry
	if err := r.db.Select(&entries, `
		SELECT id, created_at, updated_at, note
		FROM mood_entries
		WHERE created_at >= $1 AND created_at <= $2
		ORDER BY created_at ASC
	`, start, end); err != nil {
		return nil, err
	}

	return r.hydrateEntries(entries)
}

func (r *moodRepository) hydrateEntries(entries []models.MoodEntry) ([]models.MoodEntry, error) {

	if len(entries) == 0 {
		return []models.MoodEntry{}, nil
	}

	tagsByEntryID, err := r.getTagsByEntryIDs(extractMoodEntryIDs(entries))
	if err != nil {
		return nil, err
	}

	for i := range entries {
		entries[i].Tags = tagsByEntryID[entries[i].ID]
		if entries[i].Tags == nil {
			entries[i].Tags = []models.MoodTag{}
		}
	}

	return entries, nil
}

func (r *moodRepository) GetEntryByID(id string) (*models.MoodEntry, error) {
	var entry models.MoodEntry
	if err := r.db.Get(&entry, `
		SELECT id, created_at, updated_at, note
		FROM mood_entries
		WHERE id = $1
	`, id); err != nil {
		return nil, err
	}

	tagsByEntryID, err := r.getTagsByEntryIDs([]string{id})
	if err != nil {
		return nil, err
	}

	entry.Tags = tagsByEntryID[id]
	if entry.Tags == nil {
		entry.Tags = []models.MoodTag{}
	}

	return &entry, nil
}

func (r *moodRepository) UpdateEntry(id string, note *string, tagIDs []string) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	result, err := tx.Exec(`
		UPDATE mood_entries
		SET note = $2, updated_at = $3
		WHERE id = $1
	`, id, note, time.Now().UTC())
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	if _, err = tx.Exec(`DELETE FROM mood_entry_tags WHERE mood_entry_id = $1`, id); err != nil {
		return err
	}

	if err = insertMoodEntryTags(tx, id, tagIDs); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *moodRepository) getTagsByEntryIDs(entryIDs []string) (map[string][]models.MoodTag, error) {
	query, args, err := sqlx.In(`
		SELECT met.mood_entry_id,
			   mt.id,
			   mt.name,
			   mt.is_active,
			   mt.created_at,
			   mt.updated_at
		FROM mood_entry_tags met
		JOIN mood_tags mt ON mt.id = met.mood_tag_id
		WHERE met.mood_entry_id IN (?)
		ORDER BY mt.name ASC
	`, entryIDs)
	if err != nil {
		return nil, err
	}

	query = r.db.Rebind(query)

	type moodEntryTagRow struct {
		MoodEntryID string    `db:"mood_entry_id"`
		ID          string    `db:"id"`
		Name        string    `db:"name"`
		IsActive    bool      `db:"is_active"`
		CreatedAt   time.Time `db:"created_at"`
		UpdatedAt   time.Time `db:"updated_at"`
	}

	var rows []moodEntryTagRow
	if err := r.db.Select(&rows, query, args...); err != nil {
		return nil, err
	}

	result := make(map[string][]models.MoodTag, len(entryIDs))
	for _, row := range rows {
		result[row.MoodEntryID] = append(result[row.MoodEntryID], models.MoodTag{
			ID:        row.ID,
			Name:      row.Name,
			IsActive:  row.IsActive,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		})
	}

	return result, nil
}

func insertMoodEntryTags(tx *sqlx.Tx, entryID string, tagIDs []string) error {
	for _, tagID := range tagIDs {
		if _, err := tx.Exec(`
			INSERT INTO mood_entry_tags (mood_entry_id, mood_tag_id)
			VALUES ($1, $2)
		`, entryID, tagID); err != nil {
			return err
		}
	}

	return nil
}

func extractMoodEntryIDs(entries []models.MoodEntry) []string {
	ids := make([]string, 0, len(entries))
	for _, entry := range entries {
		ids = append(ids, entry.ID)
	}
	return ids
}
