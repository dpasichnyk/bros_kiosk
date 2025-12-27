package fetcher

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestICalFetcher_Name(t *testing.T) {
	f := NewICalFetcher("my-cal", "http://example.com/cal.ics")
	if f.Name() != "my-cal" {
		t.Errorf("Expected name my-cal, got %s", f.Name())
	}
}

func TestICalFetcher_Fetch(t *testing.T) {
	now := time.Now().UTC()
	future := now.Add(24 * time.Hour).Format("20060102T150405Z")
	futureEnd := now.Add(25 * time.Hour).Format("20060102T150405Z")

	mockICS := fmt.Sprintf(`BEGIN:VCALENDAR
VERSION:2.0
BEGIN:VEVENT
SUMMARY:Test Event
DTSTART:%s
DTEND:%s
LOCATION:Home
DESCRIPTION:Testing iCal
END:VEVENT
END:VCALENDAR`, future, futureEnd)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockICS))
	}))
	defer server.Close()

	f := NewICalFetcher("test-cal", server.URL)
	data, err := f.Fetch(context.Background())
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	calData, ok := data.(*CalendarData)
	if !ok {
		t.Fatalf("Expected *CalendarData, got %T", data)
	}

	if len(calData.Events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(calData.Events))
	}

	event := calData.Events[0]
	if event.Summary != "Test Event" {
		t.Errorf("Expected Summary 'Test Event', got '%s'", event.Summary)
	}
	        if event.Location != "Home" {
	                t.Errorf("Expected Location 'Home', got '%s'", event.Location)
	        }
	}
	
	func TestICalFetcher_Filtering(t *testing.T) {
	        now := time.Now().UTC()
	        format := "20060102T150405Z"
	        
	        today := now.Add(time.Hour).Format(format)
	        sixDays := now.Add(6 * 24 * time.Hour).Format(format)
	        eightDays := now.Add(8 * 24 * time.Hour).Format(format)
	
	                mockICS := fmt.Sprintf("BEGIN:VCALENDAR\r\nVERSION:2.0\r\nBEGIN:VEVENT\r\nSUMMARY:Today\r\nDTSTART:%s\r\nEND:VEVENT\r\nBEGIN:VEVENT\r\nSUMMARY:Six Days\r\nDTSTART:%s\r\nEND:VEVENT\r\nBEGIN:VEVENT\r\nSUMMARY:Eight Days\r\nDTSTART:%s\r\nEND:VEVENT\r\nEND:VCALENDAR\r\n", today, sixDays, eightDays)
	        	        server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	                w.WriteHeader(http.StatusOK)
	                w.Write([]byte(mockICS))
	        }))
	        defer server.Close()
	
	        f := NewICalFetcher("test-cal", server.URL)
	        data, err := f.Fetch(context.Background())
	        if err != nil {
	                t.Fatalf("Fetch failed: %v", err)
	        }
	
	        calData := data.(*CalendarData)
	        if len(calData.Events) != 2 {
	                t.Errorf("Expected 2 events (today and 6 days), got %d", len(calData.Events))
	        }
	
	        foundEightDays := false
	        for _, e := range calData.Events {
	                if e.Summary == "Eight Days" {
	                        foundEightDays = true
	                }
	        }
	        if foundEightDays {
	                t.Error("Expected 'Eight Days' event to be filtered out, but it was found")
	        }
	}
	
	func TestICalFetcher_SetName(t *testing.T) {	f := NewICalFetcher("old", "url")
	f.SetName("new")
	if f.Name() != "new" {
		t.Errorf("Expected name new, got %s", f.Name())
	}
}

func TestICalFetcher_FetchErrors(t *testing.T) {
	t.Run("InvalidURL", func(t *testing.T) {
		f := NewICalFetcher("test", " http://invalid")
		_, err := f.Fetch(context.Background())
		if err == nil {
			t.Error("Expected error for invalid URL, got nil")
		}
	})

	t.Run("Non200", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		f := NewICalFetcher("test", server.URL)
		_, err := f.Fetch(context.Background())
		if err == nil {
			t.Error("Expected error for 404, got nil")
		}
	})

	t.Run("InvalidICS", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("not an ics"))
		}))
		defer server.Close()

		f := NewICalFetcher("test", server.URL)
		_, err := f.Fetch(context.Background())
		if err == nil {
			t.Error("Expected error for invalid ICS, got nil")
		}
	})
}
