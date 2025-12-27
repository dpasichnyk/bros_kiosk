package fetcher

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/emersion/go-ical"
	"github.com/emersion/go-webdav/caldav"
)

// CalDAVFetcher implements the Fetcher interface for CalDAV sources.
type CalDAVFetcher struct {
	name     string
	url      string
	username string
	password string
}

// NewCalDAVFetcher creates a new instance of CalDAVFetcher.
func NewCalDAVFetcher(name, url, username, password string) *CalDAVFetcher {
	return &CalDAVFetcher{
		name:     name,
		url:      url,
		username: username,
		password: password,
	}
}

// SetName overrides the default name.
func (f *CalDAVFetcher) SetName(name string) {
	f.name = name
}

// Name returns the fetcher name.
func (f *CalDAVFetcher) Name() string {
	return f.name
}

// Fetch retrieves events from the CalDAV server.
func (f *CalDAVFetcher) Fetch(ctx context.Context) (interface{}, error) {
	// Custom transport for Basic Auth
	hc := &http.Client{
		Transport: &authTransport{
			Transport: http.DefaultTransport,
			Username:  f.username,
			Password:  f.password,
		},
		Timeout: 15 * time.Second,
	}

	client, err := caldav.NewClient(hc, f.url)
	if err != nil {
		return nil, fmt.Errorf("failed to create caldav client: %w", err)
	}

	// 1. Try to find calendars (discovery)
	calendars, err := client.FindCalendars(ctx, "")
	if err != nil {
		// If discovery fails at root, try to use the provided URL as the calendar path
		return f.fetchFromCalendar(ctx, client, f.url)
	}

	if len(calendars) == 0 {
		return f.fetchFromCalendar(ctx, client, f.url)
	}

	return f.fetchFromCalendar(ctx, client, calendars[0].Path)
}

func (f *CalDAVFetcher) fetchFromCalendar(ctx context.Context, client *caldav.Client, path string) (interface{}, error) {
	now := time.Now()
	start := now.Add(-24 * time.Hour)
	end := now.Add(30 * 24 * time.Hour)

	query := &caldav.CalendarQuery{
		CompFilter: caldav.CompFilter{
			Name: "VCALENDAR",
			Comps: []caldav.CompFilter{
				{
					Name:  "VEVENT",
					Start: start,
					End:   end,
				},
			},
		},
	}

	objs, err := client.QueryCalendar(ctx, path, query)
	if err != nil {
		return nil, fmt.Errorf("query calendar failed at %s: %w", path, err)
	}

	events := make([]CalendarEvent, 0)
	for _, obj := range objs {
		if obj.Data == nil {
			continue
		}

		for _, e := range obj.Data.Events() {
			event := CalendarEvent{}
			if prop := e.Props.Get(ical.PropSummary); prop != nil {
				event.Summary = prop.Value
			}
			if prop := e.Props.Get(ical.PropLocation); prop != nil {
				event.Location = prop.Value
			}
			if prop := e.Props.Get(ical.PropDescription); prop != nil {
				event.Description = prop.Value
			}
			if prop := e.Props.Get(ical.PropStatus); prop != nil {
				event.Status = prop.Value
			}

			if t, err := e.DateTimeStart(time.UTC); err == nil {
				event.Start = t
			}
			if t, err := e.DateTimeEnd(time.UTC); err == nil {
				event.End = t
			}

			if prop := e.Props.Get(ical.PropDateTimeStart); prop != nil {
				if prop.Params.Get("VALUE") == "DATE" {
					event.AllDay = true
				}
			}

			events = append(events, event)
		}
	}

	return &CalendarData{
		Source: f.name,
		Events: events,
	}, nil
}

type authTransport struct {
	Transport http.RoundTripper
	Username  string
	Password  string
}

func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.SetBasicAuth(t.Username, t.Password)
	return t.Transport.RoundTrip(req)
}