package scanner

import (
	"context"
	"sort"
	"testing"
)

type MockScanner struct {
	Files []string
	Err   error
}

func (m *MockScanner) Scan(ctx context.Context) ([]string, error) {
	return m.Files, m.Err
}

func TestManager_Scan(t *testing.T) {
	s1 := &MockScanner{Files: []string{"file1.jpg", "file2.png"}}
	s2 := &MockScanner{Files: []string{"file3.webp"}}

	mgr := NewManager(s1, s2)
	err := mgr.Scan(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	photos := mgr.GetPhotos()
	if len(photos) != 3 {
		t.Errorf("Expected 3 photos, got %d", len(photos))
	}

	sort.Strings(photos)
	expected := []string{"file1.jpg", "file2.png", "file3.webp"}
	for i, p := range photos {
		if p != expected[i] {
			t.Errorf("Index %d: expected %s, got %s", i, expected[i], p)
		}
	}
}
