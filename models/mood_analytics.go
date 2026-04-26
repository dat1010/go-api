package models

import "time"

type MoodAnalyticsParams struct {
	Start    time.Time
	End      time.Time
	Timezone string
	Location *time.Location
}

type MoodTagFrequency struct {
	Tag   string `json:"tag"`
	Count int    `json:"count"`
}

type MoodOverviewAnalytics struct {
	Frequency        []MoodTagFrequency `json:"frequency"`
	TotalEntries     int                `json:"totalEntries"`
	TotalTagsApplied int                `json:"totalTagsApplied"`
}

type MoodTimeOfDayPoint struct {
	Hour  int    `json:"hour"`
	Tag   string `json:"tag"`
	Count int    `json:"count"`
}

type MoodCalendarDay struct {
	Date       string `json:"date"`
	EntryCount int    `json:"entryCount"`
	Sentiment  string `json:"sentiment"`
}

type MoodPatternsAnalytics struct {
	TimeOfDay []MoodTimeOfDayPoint `json:"timeOfDay"`
	Calendar  []MoodCalendarDay    `json:"calendar"`
}

type MoodCooccurrenceNode struct {
	ID    string `json:"id"`
	Count int    `json:"count"`
}

type MoodCooccurrenceEdge struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Weight int    `json:"weight"`
}

type MoodCooccurrenceGraph struct {
	Nodes []MoodCooccurrenceNode `json:"nodes"`
	Edges []MoodCooccurrenceEdge `json:"edges"`
}

type MoodTransition struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Count int    `json:"count"`
}

type MoodInsightsAnalytics struct {
	Cooccurrence MoodCooccurrenceGraph `json:"cooccurrence"`
	Transitions  []MoodTransition      `json:"transitions"`
}
