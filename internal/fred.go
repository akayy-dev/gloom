package internal

type FREDReleaseDates struct {
	RealtimeStart string `json:"realtime_start"`
	RealtimeEnd   string `json:"realtime_end"`
	OrderBy       string `json:"order_by"`
	SortOrder     string `json:"sort_order"`
	Count         int    `json:"count"`
	Offset        int    `json:"offset"`
	Limit         int    `json:"limit"`
	ReleaseDates  []struct {
		ReleaseID   int    `json:"release_id"`
		ReleaseName string `json:"release_name"`
		Date        string `json:"date"`
	} `json:"release_dates"`
}
