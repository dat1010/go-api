package controllers

import (
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dat1010/go-api/models"
	"github.com/dat1010/go-api/services"
	"github.com/dat1010/go-api/utils"
	"github.com/gin-gonic/gin"
)

// @Summary Overview mood analytics
// @Description Return mood frequency analytics for the selected time window.
// @Tags moods
// @Produce json
// @Security Bearer
// @Param start query string false "RFC3339 or YYYY-MM-DD start"
// @Param end query string false "RFC3339 or YYYY-MM-DD end"
// @Param timezone query string false "IANA timezone, for example America/New_York"
// @Success 200 {object} models.MoodOverviewAnalytics
// @Failure 400 {object} object "Invalid query"
// @Failure 500 {object} object "Internal server error"
// @Router /moods/analytics/overview [get]
func GetMoodOverviewAnalytics(c *gin.Context) {
	params, err := buildMoodAnalyticsParams(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	auth0UserID, ok := utils.GetAuth0UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	analytics, err := moodService.GetMoodOverviewAnalytics(auth0UserID, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, analytics)
}

// @Summary Patterns mood analytics
// @Description Return time-of-day and calendar mood analytics for the selected time window.
// @Tags moods
// @Produce json
// @Security Bearer
// @Param start query string false "RFC3339 or YYYY-MM-DD start"
// @Param end query string false "RFC3339 or YYYY-MM-DD end"
// @Param timezone query string false "IANA timezone, for example America/New_York"
// @Success 200 {object} models.MoodPatternsAnalytics
// @Failure 400 {object} object "Invalid query"
// @Failure 500 {object} object "Internal server error"
// @Router /moods/analytics/patterns [get]
func GetMoodPatternsAnalytics(c *gin.Context) {
	params, err := buildMoodAnalyticsParams(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	auth0UserID, ok := utils.GetAuth0UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	analytics, err := moodService.GetMoodPatternsAnalytics(auth0UserID, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, analytics)
}

// @Summary Insights mood analytics
// @Description Return co-occurrence and transition analytics for the selected time window.
// @Tags moods
// @Produce json
// @Security Bearer
// @Param start query string false "RFC3339 or YYYY-MM-DD start"
// @Param end query string false "RFC3339 or YYYY-MM-DD end"
// @Param timezone query string false "IANA timezone, for example America/New_York"
// @Success 200 {object} models.MoodInsightsAnalytics
// @Failure 400 {object} object "Invalid query"
// @Failure 500 {object} object "Internal server error"
// @Router /moods/analytics/insights [get]
func GetMoodInsightsAnalytics(c *gin.Context) {
	params, err := buildMoodAnalyticsParams(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	auth0UserID, ok := utils.GetAuth0UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	analytics, err := moodService.GetMoodInsightsAnalytics(auth0UserID, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, analytics)
}

func buildMoodAnalyticsParams(c *gin.Context) (models.MoodAnalyticsParams, error) {
	locationName := strings.TrimSpace(c.DefaultQuery("timezone", "UTC"))
	location, err := time.LoadLocation(locationName)
	if err != nil {
		log.Printf("invalid analytics timezone %q: %v", locationName, err)
		return models.MoodAnalyticsParams{}, errors.New("timezone must be a valid IANA timezone")
	}

	now := time.Now().UTC()
	defaultStart, defaultEnd := services.DefaultMoodAnalyticsRange(now)

	start := defaultStart
	end := defaultEnd

	if rawStart := c.Query("start"); rawStart != "" {
		parsedStart, err := parseAnalyticsBoundary(rawStart, location, false)
		if err != nil {
			return models.MoodAnalyticsParams{}, err
		}
		start = parsedStart
	}

	if rawEnd := c.Query("end"); rawEnd != "" {
		parsedEnd, err := parseAnalyticsBoundary(rawEnd, location, true)
		if err != nil {
			return models.MoodAnalyticsParams{}, err
		}
		end = parsedEnd
	}

	if c.Query("start") == "" && c.Query("end") != "" {
		start = end.AddDate(0, 0, -30)
	}

	if c.Query("start") != "" && c.Query("end") == "" {
		end = now
	}

	if start.After(end) {
		return models.MoodAnalyticsParams{}, errors.New("start must be before end")
	}

	return models.MoodAnalyticsParams{
		Start:    start.UTC(),
		End:      end.UTC(),
		Timezone: locationName,
		Location: location,
	}, nil
}

func parseAnalyticsBoundary(raw string, location *time.Location, endOfDay bool) (time.Time, error) {
	if parsed, err := time.Parse(time.RFC3339, raw); err == nil {
		return parsed.UTC(), nil
	}

	parsedDate, err := time.ParseInLocation(time.DateOnly, raw, location)
	if err != nil {
		return time.Time{}, errors.New("start and end must be RFC3339 or YYYY-MM-DD")
	}

	if !endOfDay {
		return parsedDate.UTC(), nil
	}

	return parsedDate.Add(24*time.Hour - time.Nanosecond).UTC(), nil
}
