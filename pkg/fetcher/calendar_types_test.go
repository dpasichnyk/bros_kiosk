package fetcher

import (
	"encoding/json"
	"testing"
	"time"
)

func TestCalendarEventSerialization(t *testing.T) {
	now := time.Now().UTC()
	event := CalendarEvent{
		Summary:     "Meeting",
		Start:       now,
		End:         now.Add(time.Hour),
		Location:    "Room 101",
		Description: "Discuss project",
		Status:      "CONFIRMED",
		AllDay:      false,
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Failed to marshal event: %v", err)
	}

	var parsed CalendarEvent
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal event: %v", err)
	}

	if parsed.Summary != event.Summary {
		t.Errorf("Expected Summary %s, got %s", event.Summary, parsed.Summary)
	}
	if !parsed.Start.Equal(event.Start) {
		t.Errorf("Expected Start %v, got %v", event.Start, parsed.Start)
	}
	if parsed.AllDay != event.AllDay {
		t.Errorf("Expected AllDay %v, got %v", event.AllDay, parsed.AllDay)
	}
}

func TestCalendarDataStructure(t *testing.T) {
	calData := CalendarData{
		Source: "Work Calendar",
		Events: []CalendarEvent{
			{Summary: "Event 1"},
			{Summary: "Event 2"},
		},
	}

	if len(calData.Events) != 2 {
		t.Errorf("Expected 2 events, got %d", len(calData.Events))
	}
}
