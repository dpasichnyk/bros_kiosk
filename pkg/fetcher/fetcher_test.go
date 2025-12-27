package fetcher

import (
	"context"
	"testing"
	"time"
)

// TestFetcherInterface verifies that a mock implementation satisfies the Fetcher interface.
// This test implicitly defines the expected signature of the Fetcher interface.
func TestFetcherInterface(t *testing.T) {
	var _ Fetcher = &MockFetcher{}
}

// MockFetcher is a stub to verify the interface.
type MockFetcher struct{}

func (m *MockFetcher) Fetch(ctx context.Context) (interface{}, error) {
	return nil, nil
}

func (m *MockFetcher) Name() string {
	return "mock"
}

func TestStatusStruct(t *testing.T) {
	now := time.Now()
	status := Status{
		LastFetch: now,
		Error:     nil,
		IsHealthy: true,
	}

	if status.LastFetch != now {
		t.Errorf("Expected LastFetch to be %v, got %v", now, status.LastFetch)
	}
	if status.IsHealthy != true {
		t.Errorf("Expected IsHealthy to be true, got %v", status.IsHealthy)
	}
}
