package main

import (
	"strings"
	"testing"
	"time"

	"slack-rss-feed-manager/config"
	"slack-rss-feed-manager/rss"
	"slack-rss-feed-manager/state"
)

type mockSlackClient struct {
	messages []struct {
		channel string
		text    string
	}
}

func (m *mockSlackClient) PostMessage(channel, text string) error {
	if m.messages == nil {
		m.messages = make([]struct {
			channel string
			text    string
		}, 0)
	}
	m.messages = append(m.messages, struct {
		channel string
		text    string
	}{channel, text})
	return nil
}

type mockRSSClient struct {
	items []rss.FeedItem
}

func (m *mockRSSClient) FetchFeed(url string, lastUpdated time.Time) ([]rss.FeedItem, time.Time, error) {
	var latest time.Time
	for _, item := range m.items {
		if item.Published.After(latest) {
			latest = item.Published
		}
	}
	return m.items, latest, nil
}

func TestUpdateSubscriptions(t *testing.T) {
	tests := []struct {
		name            string
		config          config.Config
		initialState    state.State
		expectedFeeds   map[string]bool
		expectedChannel string
	}{
		{
			name: "adds new channel and feed",
			config: config.Config{
				Channels: []config.Channel{
					{
						SlackChannel: "test-channel",
						Feeds:        []string{"http://example.com/feed1"},
					},
				},
			},
			initialState: state.State{
				Channels: make(map[string]state.ChannelState),
			},
			expectedFeeds:   map[string]bool{"http://example.com/feed1": true},
			expectedChannel: "test-channel",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			currentState := tt.initialState
			updateSubscriptions(tt.config, &currentState)

			// Verify channel exists
			channelState, exists := currentState.Channels[tt.expectedChannel]
			if !exists {
				t.Errorf("Expected channel %s to exist", tt.expectedChannel)
				return
			}

			// Verify feeds match expected
			for feed := range channelState.Feeds {
				if !tt.expectedFeeds[feed] {
					t.Errorf("Unexpected feed found: %s", feed)
				}
			}
			for expectedFeed := range tt.expectedFeeds {
				if _, exists := channelState.Feeds[expectedFeed]; !exists {
					t.Errorf("Expected feed not found: %s", expectedFeed)
				}
			}
		})
	}
}

func TestProcessFeeds(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name            string
		config          config.Config
		initialState    state.State
		expectedPosts   int
		expectedChannel string
		mockFeedItems   []rss.FeedItem
		expectedOrder   []string
	}{
		{
			name: "posts items in chronological order (oldest first)",
			config: config.Config{
				Channels: []config.Channel{
					{
						SlackChannel: "test-channel",
						Feeds:        []string{"http://example.com/feed"},
					},
				},
			},
			initialState: state.State{
				Channels: map[string]state.ChannelState{
					"test-channel": {
						Feeds: map[string]state.FeedState{
							"http://example.com/feed": {LastUpdated: now.Add(-24 * time.Hour)},
						},
					},
				},
			},
			mockFeedItems: []rss.FeedItem{
				{
					Title:     "Newest Post",
					Link:      "http://example.com/newest",
					Published: now.Add(-1 * time.Hour),
				},
				{
					Title:     "Middle Post",
					Link:      "http://example.com/middle",
					Published: now.Add(-2 * time.Hour),
				},
				{
					Title:     "Oldest Post",
					Link:      "http://example.com/oldest",
					Published: now.Add(-3 * time.Hour),
				},
			},
			expectedPosts:   3,
			expectedChannel: "test-channel",
			expectedOrder:   []string{"Oldest Post", "Middle Post", "Newest Post"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			currentState := tt.initialState
			mockSlack := &mockSlackClient{}
			mockRSS := &mockRSSClient{items: tt.mockFeedItems}

			feedsProcessed, postsFound := processFeeds(tt.config, &currentState, mockSlack, mockRSS)

			if feedsProcessed != 1 {
				t.Errorf("expected 1 feed processed, got %d", feedsProcessed)
			}

			if postsFound != tt.expectedPosts {
				t.Errorf("expected %d posts, got %d", tt.expectedPosts, postsFound)
			}

			if len(mockSlack.messages) != tt.expectedPosts {
				t.Errorf("expected %d messages sent, got %d", tt.expectedPosts, len(mockSlack.messages))
			}

			// Verify chronological order
			if tt.expectedOrder != nil {
				for i, expectedTitle := range tt.expectedOrder {
					if i >= len(mockSlack.messages) {
						t.Errorf("missing message at position %d, expected title %s", i, expectedTitle)
						continue
					}
					if !contains(mockSlack.messages[i].text, expectedTitle) {
						t.Errorf("message at position %d has wrong title, got message %q, expected to contain %q",
							i, mockSlack.messages[i].text, expectedTitle)
					}
				}
			}
		})
	}
}

func contains(message, title string) bool {
	return strings.Contains(message, title)
}
