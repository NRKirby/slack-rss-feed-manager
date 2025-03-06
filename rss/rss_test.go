package rss

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFetchFeed(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return a sample RSS feed
		w.Header().Set("Content-Type", "application/xml")
		w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
		<feed xmlns="http://www.w3.org/2005/Atom">
			<title>Test Feed</title>
			<entry>
				<title>Test Post 1</title>
				<link href="http://example.com/post1"/>
				<updated>2024-03-01T12:00:00Z</updated>
			</entry>
			<entry>
				<title>Test Post 2</title>
				<link href="http://example.com/post2"/>
				<updated>2024-03-01T13:00:00Z</updated>
			</entry>
		</feed>`))
	}))
	defer server.Close()

	t.Run("fetch new posts", func(t *testing.T) {
		lastUpdated := time.Date(2024, 3, 1, 11, 0, 0, 0, time.UTC)
		items, newLastUpdated, err := FetchFeed(server.URL, lastUpdated)
		if err != nil {
			t.Fatalf("FetchFeed() error = %v", err)
		}

		// Should find both posts
		if len(items) != 2 {
			t.Errorf("Expected 2 items, got %d", len(items))
		}

		// Check last updated time
		expectedTime := time.Date(2024, 3, 1, 13, 0, 0, 0, time.UTC)
		if !newLastUpdated.Equal(expectedTime) {
			t.Errorf("Expected last updated %v, got %v", expectedTime, newLastUpdated)
		}
	})

	t.Run("no new posts", func(t *testing.T) {
		// Set lastUpdated to after both posts
		lastUpdated := time.Date(2024, 3, 1, 14, 0, 0, 0, time.UTC)
		items, newLastUpdated, err := FetchFeed(server.URL, lastUpdated)
		if err != nil {
			t.Fatalf("FetchFeed() error = %v", err)
		}

		if len(items) != 0 {
			t.Errorf("Expected 0 items, got %d", len(items))
		}

		if !newLastUpdated.Equal(lastUpdated) {
			t.Errorf("Expected last updated to remain %v, got %v", lastUpdated, newLastUpdated)
		}
	})
}

func TestFetchFeedErrors(t *testing.T) {
	t.Run("invalid URL", func(t *testing.T) {
		_, _, err := FetchFeed("http://invalid-url", time.Now())
		if err == nil {
			t.Error("Expected error for invalid URL, got nil")
		}
	})

	t.Run("invalid feed content", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/xml")
			w.Write([]byte(`invalid feed content`))
		}))
		defer server.Close()

		_, _, err := FetchFeed(server.URL, time.Now())
		if err == nil {
			t.Error("Expected error for invalid feed content, got nil")
		}
	})
}

func TestFormatItem(t *testing.T) {
	tests := []struct {
		name     string
		item     FeedItem
		expected string
	}{
		{
			name: "basic item",
			item: FeedItem{
				Title:     "Test Post",
				Link:      "http://example.com/post",
				FeedTitle: "Example Blog",
			},
			expected: "New post from Example Blog: Test Post\nhttp://example.com/post",
		},
		{
			name: "item with special characters",
			item: FeedItem{
				Title:     "Test & Post",
				Link:      "http://example.com/post?q=1&t=2",
				FeedTitle: "Tech & News",
			},
			expected: "New post from Tech & News: Test & Post\nhttp://example.com/post?q=1&t=2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatItem(tt.item); got != tt.expected {
				t.Errorf("FormatItem() = %v, want %v", got, tt.expected)
			}
		})
	}
}
