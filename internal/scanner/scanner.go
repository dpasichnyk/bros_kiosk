package scanner

import "context"

// Scanner defines the interface for image sources
type Scanner interface {
	Scan(ctx context.Context) ([]string, error)
}

// Supported file extensions
var SupportedExts = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".webp": true,
}
