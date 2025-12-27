package hashing

import (
	"testing"
)

func TestHash(t *testing.T) {
	data1 := map[string]interface{}{
		"id":    "weather",
		"value": "sunny",
	}
	data2 := map[string]interface{}{
		"id":    "weather",
		"value": "sunny",
	}
	data3 := map[string]interface{}{
		"id":    "weather",
		"value": "rainy",
	}

	h1, err := Hash(data1)
	if err != nil {
		t.Fatalf("Hash failed: %v", err)
	}
	h2, err := Hash(data2)
	if err != nil {
		t.Fatalf("Hash failed: %v", err)
	}
	h3, err := Hash(data3)
	if err != nil {
		t.Fatalf("Hash failed: %v", err)
	}

	if h1 != h2 {
		t.Errorf("Expected identical data to produce identical hashes, got %q and %q", h1, h2)
	}

	if h1 == h3 {
		t.Errorf("Expected different data to produce different hashes, both got %q", h1)
	}
}

func TestHash_Error(t *testing.T) {
	// Functions cannot be marshaled to JSON
	data := map[string]interface{}{
		"fn": func() {},
	}

	_, err := Hash(data)
	if err == nil {
		t.Error("Expected error for unmarshalable data, got nil")
	}
}
