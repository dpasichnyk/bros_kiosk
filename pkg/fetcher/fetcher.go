package fetcher

import (
	"context"
	"time"
)

// Status represents the metadata of the last fetch operation.
type Status struct {
	LastFetch time.Time `json:"last_fetch"`
	Error     error     `json:"-"`
	ErrorMsg  string    `json:"error,omitempty"`
	IsHealthy bool      `json:"is_healthy"`
}

// Fetcher defines the interface that all data sources must implement.
type Fetcher interface {
	// Fetch performs the actual data retrieval.
	// It respects the context for cancellation and timeouts.
	Fetch(ctx context.Context) (interface{}, error)
	
	// Name returns the unique identifier for the fetcher (e.g., "weather", "rss").
	Name() string
}

// Result wraps the actual data payload with its status.
// This is what gets pushed to the aggregation channel.
type Result struct {
	FetcherName string      `json:"fetcher_name"`
	Data        interface{} `json:"data"`
	Status      Status      `json:"status"`
}
