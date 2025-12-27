package fetcher

import (
	"context"
	"sort"
)

type CalendarAggregator struct {
	name     string
	fetchers []Fetcher
}

func NewCalendarAggregator(name string, fetchers []Fetcher) *CalendarAggregator {
	return &CalendarAggregator{
		name:     name,
		fetchers: fetchers,
	}
}

func (a *CalendarAggregator) Name() string {
	return a.name
}

func (a *CalendarAggregator) Fetch(ctx context.Context) (interface{}, error) {
	allEvents := make([]CalendarEvent, 0)
	errors := make([]error, 0)

	// Fetch from all sources
	for _, f := range a.fetchers {
		data, err := f.Fetch(ctx)
		if err != nil {
			errors = append(errors, err)
			continue
		}

		if calData, ok := data.(*CalendarData); ok {
			allEvents = append(allEvents, calData.Events...)
		}
	}

	// Sort events by start time
	sort.Slice(allEvents, func(i, j int) bool {
		return allEvents[i].Start.Before(allEvents[j].Start)
	})

	// If all failed, return error
	if len(allEvents) == 0 && len(errors) == len(a.fetchers) && len(errors) > 0 {
		return nil, errors[0]
	}

	return &CalendarData{
		Source: a.name,
		Events: allEvents,
	}, nil
}
