package fetcher

import "time"

// CalendarEvent represents a single normalized calendar event.
type CalendarEvent struct {
	Summary     string    `json:"summary"`
	Start       time.Time `json:"start"`
	End         time.Time `json:"end"`
	Location    string    `json:"location"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	AllDay      bool      `json:"all_day"`
}

// CalendarData represents the collection of events from a calendar source.
type CalendarData struct {
	Source string          `json:"source"`
	Events []CalendarEvent `json:"events"`
}
