package fetcher

import (
	"testing"
)

func TestCalDAVFetcher_Name(t *testing.T) {
	f := NewCalDAVFetcher("my-cal", "http://example.com/cal", "user", "pass")
	if f.Name() != "my-cal" {
		t.Errorf("Expected name my-cal, got %s", f.Name())
	}
}

func TestCalDAVFetcher_SetName(t *testing.T) {
	f := NewCalDAVFetcher("old", "url", "u", "p")
	f.SetName("new")
	if f.Name() != "new" {
		t.Errorf("Expected name new, got %s", f.Name())
	}
}
