package cache

import (
	"testing"
	"time"
)

func TestCache_SetGet(t *testing.T) {
	c := New()
	key := "dashboard_state"
	value := map[string]interface{}{
		"weather": "sunny",
	}

	c.Set(key, value, 30*time.Second)

	got, found := c.Get(key)
	if !found {
		t.Errorf("Expected to find key %q", key)
	}

	gotMap, ok := got.(map[string]interface{})
	if !ok {
		t.Errorf("Expected value to be map[string]interface{}, got %T", got)
	}

	if gotMap["weather"] != "sunny" {
		t.Errorf("Expected weather to be sunny, got %v", gotMap["weather"])
	}
}

func TestCache_GetNonExistent(t *testing.T) {
	c := New()
	_, found := c.Get("non_existent")
	if found {
		t.Error("Expected not to find non-existent key")
	}
}

func TestCache_Expiration(t *testing.T) {
	c := New()
	key := "short_lived"
	value := "data"
	ttl := 10 * time.Millisecond

	c.Set(key, value, ttl)

	// Verify it exists initially
	if _, found := c.Get(key); !found {
		t.Error("Expected key to exist before expiration")
	}

	// Wait for expiration
	time.Sleep(20 * time.Millisecond)

	// Verify it's gone
	if _, found := c.Get(key); found {
		t.Error("Expected key to expire")
	}
}
