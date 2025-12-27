package fetcher

import (
	"sync"
	"time"
)

// Backoff manages exponential backoff durations.
type Backoff struct {
	initial time.Duration
	max     time.Duration
	current time.Duration
	mu      sync.Mutex
}

// NewBackoff creates a new Backoff instance.
func NewBackoff(initial, max time.Duration) *Backoff {
	return &Backoff{
		initial: initial,
		max:     max,
		current: 0,
	}
}

// Next calculates the next backoff duration.
// It doubles the current duration on each call, up to the maximum.
func (b *Backoff) Next() time.Duration {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.current == 0 {
		b.current = b.initial
	} else {
		b.current *= 2
		if b.current > b.max {
			b.current = b.max
		}
	}
	return b.current
}

// Reset resets the backoff state to the initial value.
func (b *Backoff) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.current = 0
}
