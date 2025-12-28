package images

import (
	"crypto/sha256"
	"encoding/hex"
	"image"
	"os"
	"path/filepath"
	"sync"

	"github.com/disintegration/imaging"
)

type DiskCache struct {
	baseDir string
	mu      sync.RWMutex
}

func NewDiskCache(baseDir string) (*DiskCache, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, err
	}
	return &DiskCache{baseDir: baseDir}, nil
}

func (c *DiskCache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	path := c.getFilePath(key)
	if _, err := os.Stat(path); err == nil {
		return path, true
	}
	return "", false
}

func (c *DiskCache) Put(key string, img image.Image) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	path := c.getFilePath(key)

	if _, err := os.Stat(path); err == nil {
		return path, nil
	}

	if err := imaging.Save(img, path); err != nil {
		return "", err
	}

	return path, nil
}

func (c *DiskCache) getFilePath(key string) string {
	hash := sha256.Sum256([]byte(key))
	filename := hex.EncodeToString(hash[:]) + ".jpg"
	return filepath.Join(c.baseDir, filename)
}
