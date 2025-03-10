package stats

type StreamStats struct {
	TotalViews   int    `json:"total_views"`
	NewFollowers int    `json:"new_followers"`
	PeakViewers  int    `json:"peak_viewers"`
	StreamLength string `json:"stream_length"`
}

func GetWeeklyStats() StreamStats {
	// Replace this mock data with actual stats retrieval logic later
	return StreamStats{
		TotalViews:   124,
		NewFollowers: 10,
		PeakViewers:  25,
		StreamLength: "2h 45m",
	}
}
