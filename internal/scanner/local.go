package scanner

import (
	"context"
	"os"
	"path/filepath"
	"strings"
)

type LocalScanner struct {
	Path string
}

func NewLocalScanner(path string) *LocalScanner {
	return &LocalScanner{Path: path}
}

func (s *LocalScanner) Scan(ctx context.Context) ([]string, error) {
	var files []string
	err := filepath.Walk(s.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(path))
			if SupportedExts[ext] {
				files = append(files, path)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}
