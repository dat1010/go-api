package services

import (
	"testing"
	"time"

	"github.com/dat1010/go-api/models"
	"github.com/stretchr/testify/assert"
)

func TestBuildOverviewAnalyticsCountsFrequency(t *testing.T) {
	analytics := buildOverviewAnalytics([]models.MoodEntry{
		fixtureEntry("2026-03-10T10:00:00Z", "tired", "distracted"),
		fixtureEntry("2026-03-10T12:00:00Z", "tired", "sad"),
		fixtureEntry("2026-03-10T18:00:00Z", "distracted"),
	})

	assert.Equal(t, 3, analytics.TotalEntries)
	assert.Equal(t, 5, analytics.TotalTagsApplied)
	assert.Equal(t, []models.MoodTagFrequency{
		{Tag: "distracted", Count: 2},
		{Tag: "tired", Count: 2},
		{Tag: "sad", Count: 1},
	}, analytics.Frequency)
}

func TestBuildPatternsAnalyticsUsesLocalHourBucketing(t *testing.T) {
	location := mustLoadLocation(t, "America/New_York")

	analytics := buildPatternsAnalytics([]models.MoodEntry{
		fixtureEntry("2026-03-10T04:15:00Z", "anxious"),
		fixtureEntry("2026-03-10T16:30:00Z", "tired"),
		fixtureEntry("2026-03-11T01:05:00Z", "content"),
	}, location)

	assert.Equal(t, []models.MoodTimeOfDayPoint{
		{Hour: 0, Tag: "anxious", Count: 1},
		{Hour: 12, Tag: "tired", Count: 1},
		{Hour: 21, Tag: "content", Count: 1},
	}, analytics.TimeOfDay)
}

func TestBuildPatternsAnalyticsComputesCalendarSentiment(t *testing.T) {
	location := mustLoadLocation(t, "America/New_York")

	analytics := buildPatternsAnalytics([]models.MoodEntry{
		fixtureEntry("2026-03-10T14:00:00Z", "content"),
		fixtureEntry("2026-03-10T18:00:00Z", "grateful"),
		fixtureEntry("2026-03-11T04:00:00Z", "sad"),
		fixtureEntry("2026-03-11T15:00:00Z", "anxious"),
		fixtureEntry("2026-03-12T16:00:00Z", "tired"),
		fixtureEntry("2026-03-13T17:00:00Z", "distracted", "content"),
	}, location)

	assert.Equal(t, []models.MoodCalendarDay{
		{Date: "2026-03-10", EntryCount: 2, Sentiment: "positive"},
		{Date: "2026-03-11", EntryCount: 2, Sentiment: "negative"},
		{Date: "2026-03-12", EntryCount: 1, Sentiment: "neutral"},
		{Date: "2026-03-13", EntryCount: 1, Sentiment: "mixed"},
	}, analytics.Calendar)
}

func TestBuildPatternsAnalyticsTreatsCalmFocusedAndGratefulAsPositive(t *testing.T) {
	location := mustLoadLocation(t, "America/New_York")

	analytics := buildPatternsAnalytics([]models.MoodEntry{
		fixtureEntry("2026-03-12T13:00:00Z", "grateful", "content"),
		fixtureEntry("2026-03-12T16:00:00Z", "calm", "focused"),
	}, location)

	assert.Equal(t, []models.MoodCalendarDay{
		{Date: "2026-03-12", EntryCount: 2, Sentiment: "positive"},
	}, analytics.Calendar)
}

func TestBuildInsightsAnalyticsCountsCooccurrencePairs(t *testing.T) {
	analytics := buildInsightsAnalytics([]models.MoodEntry{
		fixtureEntry("2026-03-10T10:00:00Z", "sad", "lonely", "tired"),
		fixtureEntry("2026-03-10T12:00:00Z", "sad", "lonely"),
	}, time.UTC)

	assert.Equal(t, []models.MoodCooccurrenceNode{
		{ID: "lonely", Count: 2},
		{ID: "sad", Count: 2},
		{ID: "tired", Count: 1},
	}, analytics.Cooccurrence.Nodes)
	assert.Equal(t, []models.MoodCooccurrenceEdge{
		{Source: "lonely", Target: "sad", Weight: 2},
		{Source: "lonely", Target: "tired", Weight: 1},
		{Source: "sad", Target: "tired", Weight: 1},
	}, analytics.Cooccurrence.Edges)
}

func TestBuildInsightsAnalyticsCountsTransitions(t *testing.T) {
	analytics := buildInsightsAnalytics([]models.MoodEntry{
		fixtureEntry("2026-03-10T09:00:00Z", "sad"),
		fixtureEntry("2026-03-10T12:00:00Z", "distracted", "content"),
		fixtureEntry("2026-03-10T18:00:00Z", "content"),
	}, time.UTC)

	assert.Equal(t, []models.MoodTransition{
		{From: "content", To: "content", Count: 1},
		{From: "distracted", To: "content", Count: 1},
		{From: "sad", To: "content", Count: 1},
		{From: "sad", To: "distracted", Count: 1},
	}, analytics.Transitions)
}

func fixtureEntry(createdAt string, tags ...string) models.MoodEntry {
	parsed, _ := time.Parse(time.RFC3339, createdAt)
	normalizedTags := make([]models.MoodTag, 0, len(tags))
	for _, tag := range tags {
		normalizedTags = append(normalizedTags, models.MoodTag{Name: tag})
	}

	return models.MoodEntry{
		ID:        createdAt,
		CreatedAt: parsed,
		UpdatedAt: parsed,
		Tags:      normalizedTags,
	}
}

func mustLoadLocation(t *testing.T, name string) *time.Location {
	t.Helper()

	location, err := time.LoadLocation(name)
	if err != nil {
		t.Fatalf("load location %s: %v", name, err)
	}

	return location
}
