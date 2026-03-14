package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/dat1010/go-api/models"
	"github.com/dat1010/go-api/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

const testAuth0UserID = "auth0|testuser"

type mockMoodService struct {
	ListMoodTagsFunc             func() ([]models.MoodTag, error)
	CreateMoodTagFunc            func(req *models.CreateMoodTagRequest) (*models.MoodTag, error)
	ListMoodEntriesFunc          func(auth0UserID string, params models.ListMoodEntriesParams) ([]models.MoodEntry, error)
	GetMoodEntryFunc             func(auth0UserID, id string) (*models.MoodEntry, error)
	CreateMoodEntryFunc          func(auth0UserID string, req *models.CreateMoodEntryRequest) (*models.MoodEntry, error)
	UpdateMoodEntryFunc          func(auth0UserID, id string, req *models.UpdateMoodEntryRequest) (*models.MoodEntry, error)
	GetMoodOverviewAnalyticsFunc func(auth0UserID string, params models.MoodAnalyticsParams) (*models.MoodOverviewAnalytics, error)
	GetMoodPatternsAnalyticsFunc func(auth0UserID string, params models.MoodAnalyticsParams) (*models.MoodPatternsAnalytics, error)
	GetMoodInsightsAnalyticsFunc func(auth0UserID string, params models.MoodAnalyticsParams) (*models.MoodInsightsAnalytics, error)
}

func (m *mockMoodService) ListMoodTags() ([]models.MoodTag, error) {
	return m.ListMoodTagsFunc()
}

func (m *mockMoodService) CreateMoodTag(req *models.CreateMoodTagRequest) (*models.MoodTag, error) {
	return m.CreateMoodTagFunc(req)
}

func (m *mockMoodService) ListMoodEntries(auth0UserID string, params models.ListMoodEntriesParams) ([]models.MoodEntry, error) {
	return m.ListMoodEntriesFunc(auth0UserID, params)
}

func (m *mockMoodService) GetMoodEntry(auth0UserID, id string) (*models.MoodEntry, error) {
	return m.GetMoodEntryFunc(auth0UserID, id)
}

func (m *mockMoodService) CreateMoodEntry(auth0UserID string, req *models.CreateMoodEntryRequest) (*models.MoodEntry, error) {
	return m.CreateMoodEntryFunc(auth0UserID, req)
}

func (m *mockMoodService) UpdateMoodEntry(auth0UserID, id string, req *models.UpdateMoodEntryRequest) (*models.MoodEntry, error) {
	return m.UpdateMoodEntryFunc(auth0UserID, id, req)
}

func (m *mockMoodService) GetMoodOverviewAnalytics(auth0UserID string, params models.MoodAnalyticsParams) (*models.MoodOverviewAnalytics, error) {
	return m.GetMoodOverviewAnalyticsFunc(auth0UserID, params)
}

func (m *mockMoodService) GetMoodPatternsAnalytics(auth0UserID string, params models.MoodAnalyticsParams) (*models.MoodPatternsAnalytics, error) {
	return m.GetMoodPatternsAnalyticsFunc(auth0UserID, params)
}

func (m *mockMoodService) GetMoodInsightsAnalytics(auth0UserID string, params models.MoodAnalyticsParams) (*models.MoodInsightsAnalytics, error) {
	return m.GetMoodInsightsAnalyticsFunc(auth0UserID, params)
}

func addAuthenticatedUser(router *gin.Engine) {
	router.Use(func(c *gin.Context) {
		c.Set("user", validator.RegisteredClaims{Subject: testAuth0UserID})
		c.Next()
	})
}

func TestGetMoodTagsSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	moodService = &mockMoodService{
		ListMoodTagsFunc: func() ([]models.MoodTag, error) {
			return []models.MoodTag{
				{ID: "1", Name: "calm", IsActive: true},
				{ID: "2", Name: "tired", IsActive: true},
			}, nil
		},
	}

	router := gin.Default()
	router.GET("/mood-tags", GetMoodTags)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/mood-tags", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp []models.MoodTag
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Len(t, resp, 2)
	assert.Equal(t, "calm", resp[0].Name)
}

func TestCreateMoodTagConflict(t *testing.T) {
	gin.SetMode(gin.TestMode)
	moodService = &mockMoodService{
		CreateMoodTagFunc: func(req *models.CreateMoodTagRequest) (*models.MoodTag, error) {
			return nil, services.ErrMoodTagExists
		},
	}

	router := gin.Default()
	router.POST("/mood-tags", CreateMoodTag)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/mood-tags", bytes.NewBufferString(`{"name":"hopeful"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Contains(t, w.Body.String(), "mood tag already exists")
}

func TestGetMoodEntriesParsesFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	expectedFrom, _ := time.Parse(time.RFC3339, "2026-03-01T00:00:00Z")
	expectedTo, _ := time.Parse(time.RFC3339, "2026-03-10T23:59:59Z")
	moodService = &mockMoodService{
		ListMoodEntriesFunc: func(auth0UserID string, params models.ListMoodEntriesParams) ([]models.MoodEntry, error) {
			assert.Equal(t, testAuth0UserID, auth0UserID)
			assert.Equal(t, 10, params.Limit)
			assert.NotNil(t, params.From)
			assert.NotNil(t, params.To)
			if params.From == nil || params.To == nil {
				return nil, nil
			}
			assert.True(t, params.From.Equal(expectedFrom))
			assert.True(t, params.To.Equal(expectedTo))
			return []models.MoodEntry{}, nil
		},
	}

	router := gin.Default()
	addAuthenticatedUser(router)
	router.GET("/mood-entries", GetMoodEntries)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/mood-entries?limit=10&from=2026-03-01T00:00:00Z&to=2026-03-10T23:59:59Z", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "[]", string(bytes.TrimSpace(w.Body.Bytes())))
}

func TestCreateMoodEntryBadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	moodService = &mockMoodService{
		CreateMoodEntryFunc: func(auth0UserID string, req *models.CreateMoodEntryRequest) (*models.MoodEntry, error) {
			assert.Equal(t, testAuth0UserID, auth0UserID)
			return nil, services.ErrMoodTagsRequired
		},
	}

	router := gin.Default()
	addAuthenticatedUser(router)
	router.POST("/mood-entries", CreateMoodEntry)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/mood-entries", bytes.NewBufferString(`{"tagIds":[]}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "at least one tag is required")
}

func TestGetMoodEntryNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	moodService = &mockMoodService{
		GetMoodEntryFunc: func(auth0UserID, id string) (*models.MoodEntry, error) {
			assert.Equal(t, testAuth0UserID, auth0UserID)
			return nil, services.ErrMoodEntryNotFound
		},
	}

	router := gin.Default()
	addAuthenticatedUser(router)
	router.GET("/mood-entries/:id", GetMoodEntry)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/mood-entries/missing", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateMoodEntrySuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	note := "Actually felt more calm"
	moodService = &mockMoodService{
		UpdateMoodEntryFunc: func(auth0UserID, id string, req *models.UpdateMoodEntryRequest) (*models.MoodEntry, error) {
			assert.Equal(t, testAuth0UserID, auth0UserID)
			if id != "entry-1" {
				return nil, errors.New("unexpected id")
			}
			return &models.MoodEntry{
				ID:          id,
				Auth0UserID: auth0UserID,
				CreatedAt:   time.Now().UTC(),
				UpdatedAt:   time.Now().UTC(),
				Note:        &note,
				Tags:        []models.MoodTag{{ID: "tag-1", Name: "calm"}},
			}, nil
		},
	}

	router := gin.Default()
	addAuthenticatedUser(router)
	router.PUT("/mood-entries/:id", UpdateMoodEntry)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/mood-entries/entry-1", bytes.NewBufferString(`{"tagIds":["tag-1"],"note":"Actually felt more calm"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"id":"entry-1"`)
}

func TestGetMoodOverviewAnalyticsParsesTimezoneAndDates(t *testing.T) {
	gin.SetMode(gin.TestMode)
	moodService = &mockMoodService{
		GetMoodOverviewAnalyticsFunc: func(auth0UserID string, params models.MoodAnalyticsParams) (*models.MoodOverviewAnalytics, error) {
			assert.Equal(t, testAuth0UserID, auth0UserID)
			assert.Equal(t, "America/New_York", params.Timezone)
			assert.Equal(t, "2026-03-01T05:00:00Z", params.Start.Format(time.RFC3339))
			assert.Equal(t, "2026-03-11T03:59:59Z", params.End.Format(time.RFC3339))
			return &models.MoodOverviewAnalytics{}, nil
		},
	}

	router := gin.Default()
	addAuthenticatedUser(router)
	router.GET("/moods/analytics/overview", GetMoodOverviewAnalytics)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(
		http.MethodGet,
		"/moods/analytics/overview?start=2026-03-01&end=2026-03-10&timezone=America/New_York",
		nil,
	)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
