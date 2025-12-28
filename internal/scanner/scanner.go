package scanner

import "context"

type Scanner interface {
	Scan(ctx context.Context) ([]string, error)
}

var SupportedExts = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".webp": true,
}
