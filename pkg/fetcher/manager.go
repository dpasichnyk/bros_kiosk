package fetcher

import (
	"context"
	"time"
)

// FetcherConfig holds the configuration for a registered fetcher.
type FetcherConfig struct {
	Fetcher        Fetcher
	Interval       time.Duration
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
}

// Manager coordinates the execution of multiple fetchers.
type Manager struct {
	fetchers []FetcherConfig
	updates  chan Result
}

// NewManager creates a new FetcherManager.
func NewManager() *Manager {
	return &Manager{
		fetchers: make([]FetcherConfig, 0),
		updates:  make(chan Result, 20), // Buffered channel
	}
}

// Register adds a fetcher to the manager with a specific polling interval.
func (m *Manager) Register(f Fetcher, interval time.Duration) {
	m.RegisterWithBackoff(f, interval, 1*time.Second, 30*time.Second)
}

// RegisterWithBackoff adds a fetcher with custom backoff settings.
func (m *Manager) RegisterWithBackoff(f Fetcher, interval, initial, max time.Duration) {
	m.fetchers = append(m.fetchers, FetcherConfig{
		Fetcher:        f,
		Interval:       interval,
		InitialBackoff: initial,
		MaxBackoff:     max,
	})
}

// Updates returns the read-only channel for fetcher results.
func (m *Manager) Updates() <-chan Result {
	return m.updates
}

// Start begins the fetching process for all registered fetchers.
// It blocks until the context is cancelled.
func (m *Manager) Start(ctx context.Context) {
	for _, config := range m.fetchers {
		go m.runFetcher(ctx, config)
	}

	<-ctx.Done()
}

// runFetcher manages the loop for a single fetcher.
func (m *Manager) runFetcher(ctx context.Context, config FetcherConfig) {
	// Initialize backoff with configured values
	backoff := NewBackoff(config.InitialBackoff, config.MaxBackoff)

	timer := time.NewTimer(0) // Start immediately
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			// Perform Fetch
			data, err := config.Fetcher.Fetch(ctx)
			
			status := Status{
				LastFetch: time.Now(),
				Error:     err,
				IsHealthy: err == nil,
			}
			if err != nil {
				status.ErrorMsg = err.Error()
			}

			// Publish Result
			m.updates <- Result{
				FetcherName: config.Fetcher.Name(),
				Data:        data,
				Status:      status,
			}

			// Determine next run time
			if err != nil {
				// Use backoff on error
				wait := backoff.Next()
				timer.Reset(wait)
			} else {
				// Reset backoff and wait normal interval on success
				backoff.Reset()
				timer.Reset(config.Interval)
			}
		}
	}
}
