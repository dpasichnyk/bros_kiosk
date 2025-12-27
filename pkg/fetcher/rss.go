package fetcher

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"time"

	"bros_kiosk/pkg/textutil"
)

// RSSItem represents a single entry in an RSS feed.
type RSSItem struct {
	Title   string `json:"title"`
	Link    string `json:"link"`
	PubDate string `json:"pub_date"`
	Summary string `json:"summary"`
}

// RSSData represents the collection of items from a feed.
type RSSData struct {
	FeedName string    `json:"feed_name"`
	Items    []RSSItem `json:"items"`
}

// RSSFetcher implements the Fetcher interface for RSS feeds.
type RSSFetcher struct {
	name   string
	url    string
	client *http.Client
}

// NewRSSFetcher creates a new instance of RSSFetcher.
func NewRSSFetcher(name, url string) *RSSFetcher {
	return &RSSFetcher{
		name:   name,
		url:    url,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// SetName overrides the default name.
func (f *RSSFetcher) SetName(name string) {
	f.name = name
}

// Name returns the fetcher name.
func (f *RSSFetcher) Name() string {
	return f.name
}

// Fetch retrieves and parses the RSS feed.
func (f *RSSFetcher) Fetch(ctx context.Context) (interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", f.url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var rss struct {
		Channel struct {
			Items []struct {
				Title       string `xml:"title"`
				Link        string `xml:"link"`
				PubDate     string `xml:"pubDate"`
				Description string `xml:"description"`
			} `xml:"item"`
		} `xml:"channel"`
	}

	if err := xml.NewDecoder(resp.Body).Decode(&rss); err != nil {
		return nil, fmt.Errorf("failed to decode RSS: %w", err)
	}

	rawItems := rss.Channel.Items
	if len(rawItems) > 5 {
		rawItems = rawItems[:5]
	}

	// ... existing code ...

	items := make([]RSSItem, len(rawItems))
	for i, item := range rawItems {
		items[i] = RSSItem{
			Title:   item.Title,
			Link:    item.Link,
			PubDate: item.PubDate,
			Summary: textutil.CleanSummary(item.Description, item.Title),
		}
	}

	return &RSSData{
		FeedName: f.name,
		Items:    items,
	}, nil
}
