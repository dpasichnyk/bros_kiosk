package fetcher

import (
	"context"
	"errors"
	"testing"
	"time"
)

// ControllableMockFetcher allows specific behavior for testing.
type ControllableMockFetcher struct {
	name      string
	shouldErr bool
	data      interface{}
	delay     time.Duration
	callCount int
}

func (m *ControllableMockFetcher) Fetch(ctx context.Context) (interface{}, error) {
	m.callCount++
	time.Sleep(m.delay)
	if m.shouldErr {
		return nil, errors.New("simulated error")
	}
	return m.data, nil
}

func (m *ControllableMockFetcher) Name() string {
	return m.name
}

func TestManagerLifecycle(t *testing.T) {
	manager := NewManager()
	
	// Create two mock fetchers
	f1 := &ControllableMockFetcher{name: "f1", data: "data1", delay: 10 * time.Millisecond}
	f2 := &ControllableMockFetcher{name: "f2", data: "data2", delay: 10 * time.Millisecond}

	// Register them (assuming Register method or passed to New)
	manager.Register(f1, 100*time.Millisecond) // 100ms interval
	manager.Register(f2, 100*time.Millisecond)

	// Start the manager
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go manager.Start(ctx)

	// Listen for updates
	updates := manager.Updates()
	
	receivedF1 := false
	receivedF2 := false

	// Wait for a few updates
	timeout := time.After(500 * time.Millisecond)
	
	for {
		select {
		case result := <-updates:
			if result.FetcherName == "f1" {
				receivedF1 = true
			}
			if result.FetcherName == "f2" {
				receivedF2 = true
			}
			if receivedF1 && receivedF2 {
				return // Success
			}
		case <-timeout:
			t.Fatal("Timed out waiting for updates from both fetchers")
		}
	}
}

func TestManagerBackoff(t *testing.T) {
	manager := NewManager()
	
	f := &ControllableMockFetcher{name: "f1", shouldErr: true}

	// Register with very short backoff for testing
	manager.RegisterWithBackoff(f, 10*time.Millisecond, 10*time.Millisecond, 100*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go manager.Start(ctx)

	updates := manager.Updates()
	
	// Collect a few results and check timestamps
	var results []Result
	timeout := time.After(500 * time.Millisecond)

	for len(results) < 3 {
		select {
		case res := <-updates:
			results = append(results, res)
		case <-timeout:
			t.Fatalf("Timed out waiting for error results. Collected %d", len(results))
		}
	}

	// Check if interval increases (approximately)
	// 1st to 2nd: ~1s (initial backoff)
	// 2nd to 3rd: ~2s
	// This might be flaky in CI if we use real time, but let's check it's increasing.
	diff1 := results[1].Status.LastFetch.Sub(results[0].Status.LastFetch)
	diff2 := results[2].Status.LastFetch.Sub(results[1].Status.LastFetch)

	if diff2 < diff1 {
		t.Errorf("Expected backoff to increase, but %v < %v", diff2, diff1)
	}
}
