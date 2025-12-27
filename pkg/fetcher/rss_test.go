package fetcher

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRSSFetcher(t *testing.T) {
	mockRSS := `<?xml version="1.0" encoding="UTF-8" ?>
	<rss version="2.0">
	<channel>
		<title>Test Feed</title>
		<item>
			<title>Article 1</title>
			<link>http://example.com/1</link>
			<pubDate>Mon, 02 Jan 2006 15:04:05 -0700</pubDate>
		</item>
		<item>
			<title>Article 2</title>
			<link>http://example.com/2</link>
		</item>
	</channel>
	</rss>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockRSS))
	}))
	defer server.Close()

	rf := NewRSSFetcher("rss", server.URL)

	if rf.Name() != "rss" {
		t.Errorf("Expected name 'rss', got '%s'", rf.Name())
	}

	data, err := rf.Fetch(context.Background())
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	feed, ok := data.(*RSSData)
	if !ok {
		t.Fatalf("Expected *RSSData, got %T", data)
	}

	if len(feed.Items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(feed.Items))
	}
	        if feed.Items[0].Title != "Article 1" {
	                t.Errorf("Expected 'Article 1', got '%s'", feed.Items[0].Title)
	        }
	}
	
	func TestRSSFetcher_Throttling(t *testing.T) {
	        mockRSS := `<?xml version="1.0" encoding="UTF-8" ?>
	        <rss version="2.0">
	        <channel>
	                <title>Test Feed</title>
	                <item><title>1</title></item>
	                <item><title>2</title></item>
	                <item><title>3</title></item>
	                <item><title>4</title></item>
	                <item><title>5</title></item>
	                <item><title>6</title></item>
	        </channel>
	        </rss>`
	
	        server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	                w.WriteHeader(http.StatusOK)
	                w.Write([]byte(mockRSS))
	        }))
	        defer server.Close()
	
	        rf := NewRSSFetcher("rss", server.URL)
	        data, err := rf.Fetch(context.Background())
	        if err != nil {
	                t.Fatalf("Fetch failed: %v", err)
	        }
	
	        feed := data.(*RSSData)
	                if len(feed.Items) != 5 {
	                        t.Errorf("Expected 5 items, got %d", len(feed.Items))
	                }
	        }
	        
	        func TestRSSFetcher_Summary(t *testing.T) {
	                mockRSS := `<?xml version="1.0" encoding="UTF-8" ?>
	                <rss version="2.0">
	                <channel>
	                        <item>
	                                <title>Article 1</title>
	                                <description>This is a summary of article 1.</description>
	                        </item>
	                </channel>
	                </rss>`
	        
	                server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	                        w.WriteHeader(http.StatusOK)
	                        w.Write([]byte(mockRSS))
	                }))
	                defer server.Close()
	        
	                rf := NewRSSFetcher("rss", server.URL)
	                data, err := rf.Fetch(context.Background())
	                if err != nil {
	                        t.Fatalf("Fetch failed: %v", err)
	                }
	        
	                feed := data.(*RSSData)
	                if len(feed.Items) != 1 {
	                        t.Fatalf("Expected 1 item, got %d", len(feed.Items))
	                }
	                if feed.Items[0].Summary != "This is a summary of article 1." {
	                        t.Errorf("Expected summary 'This is a summary of article 1.', got '%s'", feed.Items[0].Summary)
	                }
	        }
	        
	        func TestRSSManagerIntegration(t *testing.T) {	mockRSS := `<?xml version="1.0" encoding="UTF-8" ?><rss version="2.0"><channel><item><title>News</title></item></channel></rss>`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockRSS))
	}))
	defer server.Close()

	manager := NewManager()
	rf := NewRSSFetcher("rss", server.URL)
	manager.Register(rf, 100*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go manager.Start(ctx)

	select {
	case res := <-manager.Updates():
		if res.FetcherName != "rss" {
			t.Errorf("Expected fetcher name 'rss', got '%s'", res.FetcherName)
		}
		feed := res.Data.(*RSSData)
		if feed.Items[0].Title != "News" {
			t.Errorf("Expected 'News', got '%s'", feed.Items[0].Title)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Timed out waiting for RSS update via manager")
	}
}
