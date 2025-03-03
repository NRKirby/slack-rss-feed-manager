package main

import (
	"testing"
	"time"

	"slack-rss-feed-manager/config"
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
	oldTime := time.Now().Add(-24 * time.Hour)
	tests := []struct {
		name            string
		config          config.Config
		initialState    state.State
		expectedPosts   int
		expectedChannel string
	}{
		{
			name: "processes new feed items",
			config: config.Config{
				Channels: []config.Channel{
					{
						SlackChannel: "test-channel",
						Feeds:        []string{"http://example.com/feed1"},
					},
				},
			},
			initialState: state.State{
				Channels: map[string]state.ChannelState{
					"test-channel": {
						Feeds: map[string]state.FeedState{
							"http://example.com/feed1": {LastUpdated: oldTime},
						},
					},
				},
			},
			expectedPosts:   0, // Since we can't actually fetch feeds in tests
			expectedChannel: "test-channel",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			currentState := tt.initialState
			mockSlack := &mockSlackClient{}

			feedsProcessed, postsFound := processFeeds(tt.config, &currentState, mockSlack)

			// Verify number of feeds processed
			expectedFeeds := len(tt.config.Channels[0].Feeds)
			if feedsProcessed != expectedFeeds {
				t.Errorf("Expected to process %d feeds, got %d", expectedFeeds, feedsProcessed)
			}

			// Verify posts found
			if postsFound != tt.expectedPosts {
				t.Errorf("Expected to find %d posts, got %d", tt.expectedPosts, postsFound)
			}

			// Verify messages sent to correct channel
			for _, msg := range mockSlack.messages {
				if msg.channel != "#"+tt.expectedChannel {
					t.Errorf("Expected message to channel #%s, got %s", tt.expectedChannel, msg.channel)
				}
			}
		})
	}
}
