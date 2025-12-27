package scanner

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestScanLocal(t *testing.T) {
	// Create temp dir structure
	tmpDir, err := os.MkdirTemp("", "photos")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create valid files
	createFile(t, tmpDir, "photo1.jpg")
	createFile(t, tmpDir, "photo2.PNG")
	createFile(t, tmpDir, "nested/photo3.webp")

	// Create invalid files
	createFile(t, tmpDir, "doc.txt")
	createFile(t, tmpDir, "nested/script.js")

	scanner := NewLocalScanner(tmpDir)
	files, err := scanner.Scan(context.Background())
	if err != nil {
		t.Fatalf("ScanLocal failed: %v", err)
	}

	expected := []string{
		filepath.Join(tmpDir, "nested/photo3.webp"),
		filepath.Join(tmpDir, "photo1.jpg"),
		filepath.Join(tmpDir, "photo2.PNG"),
	}

	sort.Strings(files)
	sort.Strings(expected)

	if len(files) != len(expected) {
		t.Errorf("Expected %d files, got %d: %v", len(expected), len(files), files)
		return
	}

	for i := range files {
		if files[i] != expected[i] {
			t.Errorf("Index %d: expected %s, got %s", i, expected[i], files[i])
		}
	}
}

func createFile(t *testing.T, base, path string) {
	fullPath := filepath.Join(base, path)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(fullPath, []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}
}
