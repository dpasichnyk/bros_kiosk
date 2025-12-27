package fetcher

import (
	"testing"
	"time"
)

func TestBackoff(t *testing.T) {
	initial := 1 * time.Second
	max := 10 * time.Second

	b := NewBackoff(initial, max)

	// Test 1: Initial call should return 0 (no wait for first attempt) or initial? 
	// Usually backoff is calculated AFTER a failure. 
	// Let's assume Next() is called after a failure to get the wait duration.
	
	// First failure
	wait1 := b.Next()
	if wait1 != initial {
		t.Errorf("Expected first backoff to be %v, got %v", initial, wait1)
	}

	// Second failure (should double)
	wait2 := b.Next()
	if wait2 != 2*time.Second {
		t.Errorf("Expected second backoff to be 2s, got %v", wait2)
	}

	// Third failure (should be 4s)
	wait3 := b.Next()
	if wait3 != 4*time.Second {
		t.Errorf("Expected third backoff to be 4s, got %v", wait3)
	}

	// Fourth failure (should be 8s)
	wait4 := b.Next()
	if wait4 != 8*time.Second {
		t.Errorf("Expected fourth backoff to be 8s, got %v", wait4)
	}

	// Fifth failure (should be capped at 10s)
	wait5 := b.Next()
	if wait5 != max {
		t.Errorf("Expected fifth backoff to be capped at %v, got %v", max, wait5)
	}

	// Test Reset
	b.Reset()
	waitAfterReset := b.Next()
	if waitAfterReset != initial {
		t.Errorf("Expected backoff after reset to be %v, got %v", initial, waitAfterReset)
	}
}
