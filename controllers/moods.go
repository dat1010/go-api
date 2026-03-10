package controllers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/dat1010/go-api/models"
	"github.com/dat1010/go-api/services"
	"github.com/gin-gonic/gin"
)

var moodService services.MoodService

func SetMoodService(service services.MoodService) {
	moodService = service
}

// @Summary List mood tags
// @Description Return all active mood tags.
// @Tags moods
// @Produce json
// @Success 200 {array} models.MoodTag
// @Failure 500 {object} object "Internal server error"
// @Router /mood-tags [get]
func GetMoodTags(c *gin.Context) {
	tags, err := moodService.ListMoodTags()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tags)
}

// @Summary Create mood tag
// @Description Create a custom mood tag.
// @Tags moods
// @Accept json
// @Produce json
// @Param request body models.CreateMoodTagRequest true "Mood tag payload"
// @Success 201 {object} models.MoodTag
// @Failure 400 {object} object "Invalid request"
// @Failure 409 {object} object "Duplicate tag"
// @Failure 500 {object} object "Internal server error"
// @Router /mood-tags [post]
func CreateMoodTag(c *gin.Context) {
	var req models.CreateMoodTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tag, err := moodService.CreateMoodTag(&req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrMoodTagNameRequired),
			err.Error() == "tag name must be 50 characters or fewer":
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, services.ErrMoodTagExists):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, tag)
}

// @Summary List mood entries
// @Description Return mood entry history in newest-first order.
// @Tags moods
// @Produce json
// @Param limit query int false "Maximum number of entries to return"
// @Param from query string false "RFC3339 lower bound for createdAt"
// @Param to query string false "RFC3339 upper bound for createdAt"
// @Success 200 {array} models.MoodEntry
// @Failure 400 {object} object "Invalid query"
// @Failure 500 {object} object "Internal server error"
// @Router /mood-entries [get]
func GetMoodEntries(c *gin.Context) {
	params, err := buildMoodEntriesParams(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entries, err := moodService.ListMoodEntries(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, entries)
}

// @Summary Get mood entry
// @Description Return a single hydrated mood entry.
// @Tags moods
// @Produce json
// @Param id path string true "Mood entry ID"
// @Success 200 {object} models.MoodEntry
// @Failure 400 {object} object "Invalid ID"
// @Failure 404 {object} object "Mood entry not found"
// @Router /mood-entries/{id} [get]
func GetMoodEntry(c *gin.Context) {
	entry, err := moodService.GetMoodEntry(c.Param("id"))
	if err != nil {
		switch {
		case errors.Is(err, services.ErrMoodEntryNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, entry)
}

// @Summary Create mood entry
// @Description Create a mood entry with one or more selected tags.
// @Tags moods
// @Accept json
// @Produce json
// @Param request body models.CreateMoodEntryRequest true "Mood entry payload"
// @Success 201 {object} models.MoodEntry
// @Failure 400 {object} object "Invalid request"
// @Failure 500 {object} object "Internal server error"
// @Router /mood-entries [post]
func CreateMoodEntry(c *gin.Context) {
	var req models.CreateMoodEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entry, err := moodService.CreateMoodEntry(&req)
	if err != nil {
		handleMoodEntryError(c, err)
		return
	}

	c.JSON(http.StatusCreated, entry)
}

// @Summary Update mood entry
// @Description Replace the tag set and note for an existing mood entry.
// @Tags moods
// @Accept json
// @Produce json
// @Param id path string true "Mood entry ID"
// @Param request body models.UpdateMoodEntryRequest true "Mood entry payload"
// @Success 200 {object} models.MoodEntry
// @Failure 400 {object} object "Invalid request"
// @Failure 404 {object} object "Mood entry not found"
// @Failure 500 {object} object "Internal server error"
// @Router /mood-entries/{id} [put]
func UpdateMoodEntry(c *gin.Context) {
	var req models.UpdateMoodEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entry, err := moodService.UpdateMoodEntry(c.Param("id"), &req)
	if err != nil {
		handleMoodEntryError(c, err)
		return
	}

	c.JSON(http.StatusOK, entry)
}

func handleMoodEntryError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrMoodEntryNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, services.ErrMoodTagsRequired),
		errors.Is(err, services.ErrInvalidMoodTagIDs),
		err.Error() == "note must be 2000 characters or fewer":
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func buildMoodEntriesParams(c *gin.Context) (models.ListMoodEntriesParams, error) {
	params := models.ListMoodEntriesParams{Limit: 30}

	if rawLimit := c.Query("limit"); rawLimit != "" {
		limit, err := strconv.Atoi(rawLimit)
		if err != nil {
			return params, errors.New("limit must be an integer")
		}
		params.Limit = limit
	}

	if rawFrom := c.Query("from"); rawFrom != "" {
		from, err := time.Parse(time.RFC3339, rawFrom)
		if err != nil {
			return params, errors.New("from must be RFC3339")
		}
		params.From = &from
	}

	if rawTo := c.Query("to"); rawTo != "" {
		to, err := time.Parse(time.RFC3339, rawTo)
		if err != nil {
			return params, errors.New("to must be RFC3339")
		}
		params.To = &to
	}

	return params, nil
}
