package main

import (
	"log"
	"os"
	"time"

	"slack-rss-feed-manager/config"
	"slack-rss-feed-manager/rss"
	"slack-rss-feed-manager/slack"
	st "slack-rss-feed-manager/state"
)

func main() {
	token := os.Getenv("SLACK_BOT_TOKEN")
	if token == "" {
		log.Fatal("SLACK_BOT_TOKEN not set")
	}
	slackClient := slack.NewClient(token)

	configFile := "config.yaml"
	stateFile := "state.json"

	// Load config
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Load state
	currentState, err := st.LoadState(stateFile)
	if err != nil {
		log.Fatalf("Failed to load state: %v", err)
	}

	// Update subscriptions and process feeds
	updateSubscriptions(cfg, &currentState)
	processFeeds(cfg, &currentState, slackClient)

	// Save updated state
	if err := currentState.Save(stateFile); err != nil {
		log.Fatalf("Failed to save state: %v", err)
	}
}

func updateSubscriptions(cfg config.Config, state *st.State) {
	for _, ch := range cfg.Channels {
		if _, ok := state.Channels[ch.SlackChannel]; !ok {
			state.Channels[ch.SlackChannel] = st.ChannelState{Feeds: make(map[string]st.FeedState)}
		}
		channelState := state.Channels[ch.SlackChannel]
		// Add new feeds
		for _, feed := range ch.Feeds {
			if _, ok := channelState.Feeds[feed]; !ok {
				channelState.Feeds[feed] = st.FeedState{LastUpdated: time.Now()}
			}
		}
		// Remove feeds not in config (optional, for full state management)
		for feed := range channelState.Feeds {
			found := false
			for _, f := range ch.Feeds {
				if f == feed {
					found = true
					break
				}
			}
			if !found {
				delete(channelState.Feeds, feed)
			}
		}
		state.Channels[ch.SlackChannel] = channelState
	}
}

func processFeeds(cfg config.Config, state *st.State, slackClient *slack.Client) {
	for _, ch := range cfg.Channels {
		channel := ch.SlackChannel
		chState := state.Channels[channel]
		for _, feedURL := range ch.Feeds {
			lastUpdated := chState.Feeds[feedURL].LastUpdated
			items, newLastUpdated, err := rss.FetchFeed(feedURL, lastUpdated)
			if err != nil {
				log.Printf("Error fetching feed %s: %v", feedURL, err)
				continue
			}
			for _, item := range items {
				if err := slackClient.PostMessage("#"+channel, rss.FormatItem(item)); err != nil {
					log.Printf("Error posting to Slack: %v", err)
				}
			}
			if !newLastUpdated.Equal(lastUpdated) {
				chState.Feeds[feedURL] = st.FeedState{LastUpdated: newLastUpdated}
				state.Channels[channel] = chState
			}
		}
	}
}
