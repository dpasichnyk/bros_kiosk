package assets

import "embed"

// FS is the embedded filesystem containing all assets
//go:embed all:static all:templates
var FS embed.FS
