package fetcher

import (
	"context"
	"fmt"
	"net/http"
	"time"

	ical "github.com/arran4/golang-ical"
)

// ICalFetcher implements the Fetcher interface for iCal/ICS feeds.
type ICalFetcher struct {
	name   string
	url    string
	client *http.Client
}

// NewICalFetcher creates a new instance of ICalFetcher.
func NewICalFetcher(name, url string) *ICalFetcher {
	return &ICalFetcher{
		name: name,
		url:  url,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SetName overrides the default name.
func (f *ICalFetcher) SetName(name string) {
	f.name = name
}

// Name returns the fetcher name.
func (f *ICalFetcher) Name() string {
	return f.name
}

// Fetch retrieves and parses the iCal feed.
func (f *ICalFetcher) Fetch(ctx context.Context) (interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", f.url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	cal, err := ical.ParseCalendar(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse calendar: %w", err)
	}

	now := time.Now()
	sevenDaysFromNow := now.Add(7 * 24 * time.Hour)

	events := make([]CalendarEvent, 0)
	for _, vEvent := range cal.Events() {
		event := CalendarEvent{}

		if prop := vEvent.GetProperty(ical.ComponentPropertySummary); prop != nil {
			event.Summary = prop.Value
		}
		if prop := vEvent.GetProperty(ical.ComponentPropertyLocation); prop != nil {
			event.Location = prop.Value
		}
		if prop := vEvent.GetProperty(ical.ComponentPropertyDescription); prop != nil {
			event.Description = prop.Value
		}
		if prop := vEvent.GetProperty(ical.ComponentPropertyStatus); prop != nil {
			event.Status = prop.Value
		}

		if t, err := vEvent.GetStartAt(); err == nil {
			event.Start = t
		}
		if t, err := vEvent.GetEndAt(); err == nil {
			event.End = t
		}

		// Skip events outside the next 7 days
		// If Start is not set, we might want to skip it or include it.
		// Usually we want to skip it if we can't determine when it starts.
		if event.Start.IsZero() {
			continue
		}

		// If the event ended more than an hour ago, skip it
		if !event.End.IsZero() && event.End.Before(now.Add(-1*time.Hour)) {
			continue
		}

		// If the event starts more than 7 days from now, skip it
		if event.Start.After(sevenDaysFromNow) {
			continue
		}

		// Simple all-day detection: if GetAllDayStartAt doesn't error
		if _, err := vEvent.GetAllDayStartAt(); err == nil {
			event.AllDay = true
		}

		events = append(events, event)
	}

	return &CalendarData{
		Source: f.name,
		Events: events,
	}, nil
}