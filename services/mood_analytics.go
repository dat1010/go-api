package services

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/dat1010/go-api/models"
)

const analyticsDefaultWindowDays = 30

type sentimentCategory string

const (
	sentimentPositive sentimentCategory = "positive"
	sentimentNeutral  sentimentCategory = "neutral"
	sentimentNegative sentimentCategory = "negative"
	sentimentMixed    sentimentCategory = "mixed"
	sentimentUnknown  sentimentCategory = "unknown"
)

var moodSentimentCategories = map[string]sentimentCategory{
	"content":    sentimentPositive,
	"tired":      sentimentNeutral,
	"distracted": sentimentNeutral,
	"sad":        sentimentNegative,
	"anxious":    sentimentNegative,
	"frustrated": sentimentNegative,
	"lonely":     sentimentNegative,
	"numb":       sentimentNegative,
	"stressed":   sentimentNegative,
}

func (s *moodService) GetMoodOverviewAnalytics(params models.MoodAnalyticsParams) (*models.MoodOverviewAnalytics, error) {
	entries, err := s.repo.ListEntriesForAnalytics(params.Start, params.End)
	if err != nil {
		return nil, err
	}

	return buildOverviewAnalytics(entries), nil
}

func (s *moodService) GetMoodPatternsAnalytics(params models.MoodAnalyticsParams) (*models.MoodPatternsAnalytics, error) {
	entries, err := s.repo.ListEntriesForAnalytics(params.Start, params.End)
	if err != nil {
		return nil, err
	}

	return buildPatternsAnalytics(entries, params.Location), nil
}

func (s *moodService) GetMoodInsightsAnalytics(params models.MoodAnalyticsParams) (*models.MoodInsightsAnalytics, error) {
	entries, err := s.repo.ListEntriesForAnalytics(params.Start, params.End)
	if err != nil {
		return nil, err
	}

	return buildInsightsAnalytics(entries, params.Location), nil
}

func buildOverviewAnalytics(entries []models.MoodEntry) *models.MoodOverviewAnalytics {
	counts := make(map[string]int)
	totalTagsApplied := 0

	for _, entry := range entries {
		for _, tagName := range extractNormalizedTagNames(entry.Tags) {
			counts[tagName]++
			totalTagsApplied++
		}
	}

	return &models.MoodOverviewAnalytics{
		Frequency:        sortFrequency(counts),
		TotalEntries:     len(entries),
		TotalTagsApplied: totalTagsApplied,
	}
}

func buildPatternsAnalytics(entries []models.MoodEntry, location *time.Location) *models.MoodPatternsAnalytics {
	if location == nil {
		location = time.UTC
	}

	timeOfDayCounts := make(map[string]int)
	calendarBuckets := make(map[string]*calendarAggregation)

	for _, entry := range entries {
		localTime := entry.CreatedAt.In(location)
		tagNames := extractNormalizedTagNames(entry.Tags)

		for _, tagName := range tagNames {
			timeOfDayCounts[timeOfDayKey(localTime.Hour(), tagName)]++
		}

		dateKey := localTime.Format(time.DateOnly)
		bucket := calendarBuckets[dateKey]
		if bucket == nil {
			bucket = &calendarAggregation{}
			bucket.ensureSentiments()
			calendarBuckets[dateKey] = bucket
		}

		bucket.entryCount++
		for _, tagName := range tagNames {
			bucket.sentiments[classifyMoodTag(tagName)] = true
		}
	}

	return &models.MoodPatternsAnalytics{
		TimeOfDay: sortTimeOfDayPoints(timeOfDayCounts),
		Calendar:  sortCalendarDays(calendarBuckets),
	}
}

func buildInsightsAnalytics(entries []models.MoodEntry, location *time.Location) *models.MoodInsightsAnalytics {
	_ = location

	nodeCounts := make(map[string]int)
	edgeCounts := make(map[string]int)
	transitionCounts := make(map[string]int)

	sortedEntries := append([]models.MoodEntry(nil), entries...)
	sort.Slice(sortedEntries, func(i, j int) bool {
		return sortedEntries[i].CreatedAt.Before(sortedEntries[j].CreatedAt)
	})

	uniqueTagsByEntry := make([][]string, 0, len(sortedEntries))

	for _, entry := range sortedEntries {
		tagNames := extractNormalizedTagNames(entry.Tags)
		uniqueTagsByEntry = append(uniqueTagsByEntry, tagNames)

		for _, tagName := range tagNames {
			nodeCounts[tagName]++
		}

		for _, pair := range generatePairKeys(tagNames) {
			edgeCounts[pair]++
		}
	}

	for i := 1; i < len(uniqueTagsByEntry); i++ {
		for _, from := range uniqueTagsByEntry[i-1] {
			for _, to := range uniqueTagsByEntry[i] {
				transitionCounts[from+"->"+to]++
			}
		}
	}

	return &models.MoodInsightsAnalytics{
		Cooccurrence: models.MoodCooccurrenceGraph{
			Nodes: sortCooccurrenceNodes(nodeCounts),
			Edges: sortCooccurrenceEdges(edgeCounts),
		},
		Transitions: sortTransitions(transitionCounts),
	}
}

func DefaultMoodAnalyticsRange(now time.Time) (time.Time, time.Time) {
	end := now.UTC()
	start := end.AddDate(0, 0, -analyticsDefaultWindowDays)
	return start, end
}

func extractNormalizedTagNames(tags []models.MoodTag) []string {
	seen := make(map[string]struct{}, len(tags))
	names := make([]string, 0, len(tags))

	for _, tag := range tags {
		name := strings.ToLower(strings.TrimSpace(tag.Name))
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		names = append(names, name)
	}

	sort.Strings(names)
	return names
}

func sortFrequency(counts map[string]int) []models.MoodTagFrequency {
	frequency := make([]models.MoodTagFrequency, 0, len(counts))
	for tagName, count := range counts {
		frequency = append(frequency, models.MoodTagFrequency{Tag: tagName, Count: count})
	}

	sort.Slice(frequency, func(i, j int) bool {
		if frequency[i].Count == frequency[j].Count {
			return frequency[i].Tag < frequency[j].Tag
		}
		return frequency[i].Count > frequency[j].Count
	})

	return frequency
}

type calendarAggregation struct {
	entryCount int
	sentiments map[sentimentCategory]bool
}

func (c *calendarAggregation) ensureSentiments() {
	if c.sentiments == nil {
		c.sentiments = make(map[sentimentCategory]bool)
	}
}

func sortTimeOfDayPoints(counts map[string]int) []models.MoodTimeOfDayPoint {
	points := make([]models.MoodTimeOfDayPoint, 0, len(counts))
	for key, count := range counts {
		hour, tagName := parseTimeOfDayKey(key)
		points = append(points, models.MoodTimeOfDayPoint{
			Hour:  hour,
			Tag:   tagName,
			Count: count,
		})
	}

	sort.Slice(points, func(i, j int) bool {
		if points[i].Hour == points[j].Hour {
			if points[i].Count == points[j].Count {
				return points[i].Tag < points[j].Tag
			}
			return points[i].Count > points[j].Count
		}
		return points[i].Hour < points[j].Hour
	})

	return points
}

func sortCalendarDays(buckets map[string]*calendarAggregation) []models.MoodCalendarDay {
	days := make([]models.MoodCalendarDay, 0, len(buckets))
	for dateKey, bucket := range buckets {
		days = append(days, models.MoodCalendarDay{
			Date:       dateKey,
			EntryCount: bucket.entryCount,
			Sentiment:  string(resolveCalendarSentiment(bucket.sentiments)),
		})
	}

	sort.Slice(days, func(i, j int) bool {
		return days[i].Date < days[j].Date
	})

	return days
}

func sortCooccurrenceNodes(counts map[string]int) []models.MoodCooccurrenceNode {
	nodes := make([]models.MoodCooccurrenceNode, 0, len(counts))
	for id, count := range counts {
		nodes = append(nodes, models.MoodCooccurrenceNode{ID: id, Count: count})
	}

	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].Count == nodes[j].Count {
			return nodes[i].ID < nodes[j].ID
		}
		return nodes[i].Count > nodes[j].Count
	})

	return nodes
}

func sortCooccurrenceEdges(counts map[string]int) []models.MoodCooccurrenceEdge {
	edges := make([]models.MoodCooccurrenceEdge, 0, len(counts))
	for key, weight := range counts {
		source, target := parsePairKey(key)
		edges = append(edges, models.MoodCooccurrenceEdge{
			Source: source,
			Target: target,
			Weight: weight,
		})
	}

	sort.Slice(edges, func(i, j int) bool {
		if edges[i].Weight == edges[j].Weight {
			if edges[i].Source == edges[j].Source {
				return edges[i].Target < edges[j].Target
			}
			return edges[i].Source < edges[j].Source
		}
		return edges[i].Weight > edges[j].Weight
	})

	return edges
}

func sortTransitions(counts map[string]int) []models.MoodTransition {
	transitions := make([]models.MoodTransition, 0, len(counts))
	for key, count := range counts {
		from, to := parseTransitionKey(key)
		transitions = append(transitions, models.MoodTransition{
			From:  from,
			To:    to,
			Count: count,
		})
	}

	sort.Slice(transitions, func(i, j int) bool {
		if transitions[i].Count == transitions[j].Count {
			if transitions[i].From == transitions[j].From {
				return transitions[i].To < transitions[j].To
			}
			return transitions[i].From < transitions[j].From
		}
		return transitions[i].Count > transitions[j].Count
	})

	return transitions
}

func classifyMoodTag(tagName string) sentimentCategory {
	if category, ok := moodSentimentCategories[tagName]; ok {
		return category
	}
	return sentimentUnknown
}

func resolveCalendarSentiment(sentiments map[sentimentCategory]bool) sentimentCategory {
	if len(sentiments) == 0 {
		return sentimentMixed
	}

	hasPositive := sentiments[sentimentPositive]
	hasNeutral := sentiments[sentimentNeutral]
	hasNegative := sentiments[sentimentNegative]
	hasUnknown := sentiments[sentimentUnknown]

	switch {
	case hasPositive && !hasNeutral && !hasNegative && !hasUnknown:
		return sentimentPositive
	case hasNegative && !hasPositive && !hasNeutral && !hasUnknown:
		return sentimentNegative
	case hasNeutral && !hasPositive && !hasNegative && !hasUnknown:
		return sentimentNeutral
	default:
		return sentimentMixed
	}
}

func generatePairKeys(tagNames []string) []string {
	pairs := make([]string, 0)
	for i := 0; i < len(tagNames); i++ {
		for j := i + 1; j < len(tagNames); j++ {
			pairs = append(pairs, tagNames[i]+"|"+tagNames[j])
		}
	}
	return pairs
}

func timeOfDayKey(hour int, tagName string) string {
	return strings.Join([]string{timeHourString(hour), tagName}, "|")
}

func parseTimeOfDayKey(key string) (int, string) {
	parts := strings.SplitN(key, "|", 2)
	return parseTimeHour(parts[0]), parts[1]
}

func parsePairKey(key string) (string, string) {
	parts := strings.SplitN(key, "|", 2)
	return parts[0], parts[1]
}

func parseTransitionKey(key string) (string, string) {
	parts := strings.SplitN(key, "->", 2)
	return parts[0], parts[1]
}

func timeHourString(hour int) string {
	return fmt.Sprintf("%02d", hour)
}

func parseTimeHour(hour string) int {
	parsed, _ := time.Parse("15", hour)
	return parsed.Hour()
}
