package scanner

import (
	"context"
	"sync"
)

type Manager struct {
	scanners []Scanner
	photos   []string
	mu       sync.RWMutex
}

func NewManager(scanners ...Scanner) *Manager {
	return &Manager{
		scanners: scanners,
		photos:   make([]string, 0),
	}
}

func (m *Manager) Scan(ctx context.Context) error {
	var allPhotos []string

	type result struct {
		files []string
		err   error
	}

	ch := make(chan result, len(m.scanners))
	var wg sync.WaitGroup

	for _, s := range m.scanners {
		wg.Add(1)
		go func(sc Scanner) {
			defer wg.Done()
			files, err := sc.Scan(ctx)
			ch <- result{files: files, err: err}
		}(s)
	}

	wg.Wait()
	close(ch)

	for res := range ch {
		if res.err != nil {
			// For now, return the first error.
			// In production, we might want to log this and continue with partial results.
			return res.err
		}
		allPhotos = append(allPhotos, res.files...)
	}

	m.mu.Lock()
	m.photos = allPhotos
	m.mu.Unlock()
	return nil
}

func (m *Manager) GetPhotos() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	dst := make([]string, len(m.photos))
	copy(dst, m.photos)
	return dst
}
