package hashing

import (
	"reflect"
	"testing"
)

func TestDiff(t *testing.T) {
	oldState := map[string]interface{}{
		"weather": "sunny",
		"news":    "all good",
	}
	newState := map[string]interface{}{
		"weather": "rainy",
		"news":    "all good",
	}
	newState2 := map[string]interface{}{
		"weather": "rainy",
		"news":    "breaking news",
	}

	tests := []struct {
		name     string
		old      map[string]interface{}
		new      map[string]interface{}
		expected []string
	}{
		{
			name:     "one changed",
			old:      oldState,
			new:      newState,
			expected: []string{"weather"},
		},
		{
			name:     "two changed",
			old:      oldState,
			new:      newState2,
			expected: []string{"news", "weather"}, // sorted alphabetically
		},
		{
			name:     "none changed",
			old:      oldState,
			new:      oldState,
			expected: []string{},
		},
		{
			name:     "new section added",
			old:      oldState,
			new:      map[string]interface{}{"weather": "sunny", "news": "all good", "calendar": "busy"},
			expected: []string{"calendar"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Diff(tt.old, tt.new)
			if err != nil {
				t.Fatalf("Diff failed: %v", err)
			}
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("Diff() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestDiff_Error(t *testing.T) {
	oldState := map[string]interface{}{
		"fn": func() {},
	}
	newState := map[string]interface{}{
		"fn": "something else",
	}

	_, err := Diff(oldState, newState)
	if err == nil {
		t.Error("Expected error for unmarshalable data in oldState, got nil")
	}

	newState2 := map[string]interface{}{
		"fn": func() {},
	}
	oldState2 := map[string]interface{}{
		"fn": "something",
	}

	_, err = Diff(oldState2, newState2)
	if err == nil {
		t.Error("Expected error for unmarshalable data in newState, got nil")
	}
}
