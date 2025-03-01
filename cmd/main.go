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
	startTime := time.Now()
	log.Printf("RSS Feed Manager starting at %s", startTime.Format(time.RFC3339))

	token := os.Getenv("SLACK_BOT_TOKEN")
	if token == "" {
		log.Fatal("SLACK_BOT_TOKEN not set")
	}
	slackClient := slack.NewClient(token)
	log.Printf("Slack client initialized")

	configFile := "config.yaml"
	stateFile := "state.json"

	// Load config
	log.Printf("Loading config from %s", configFile)
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	log.Printf("Config loaded successfully: monitoring %d channels", len(cfg.Channels))

	// Load state
	log.Printf("Loading state from %s", stateFile)
	currentState, err := st.LoadState(stateFile)
	if err != nil {
		log.Fatalf("Failed to load state: %v", err)
	}
	log.Printf("State loaded successfully")

	// Update subscriptions and process feeds
	log.Printf("Updating subscriptions...")
	updateSubscriptions(cfg, &currentState)
	log.Printf("Processing feeds...")
	feedsProcessed, postsFound := processFeeds(cfg, &currentState, slackClient)

	// Save updated state
	log.Printf("Saving updated state to %s", stateFile)
	if err := currentState.Save(stateFile); err != nil {
		log.Fatalf("Failed to save state: %v", err)
	}

	duration := time.Since(startTime)
	log.Printf("RSS Feed Manager completed successfully in %v", duration)
	log.Printf("Summary: Processed %d feeds, found %d new posts", feedsProcessed, postsFound)
}

func updateSubscriptions(cfg config.Config, state *st.State) {
	for _, ch := range cfg.Channels {
		if _, ok := state.Channels[ch.SlackChannel]; !ok {
			log.Printf("Adding new channel to state: %s", ch.SlackChannel)
			state.Channels[ch.SlackChannel] = st.ChannelState{Feeds: make(map[string]st.FeedState)}
		}
		channelState := state.Channels[ch.SlackChannel]
		// Add new feeds
		for _, feed := range ch.Feeds {
			if _, ok := channelState.Feeds[feed]; !ok {
				log.Printf("Adding new feed to channel %s: %s", ch.SlackChannel, feed)
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
				log.Printf("Removing feed from channel %s: %s", ch.SlackChannel, feed)
				delete(channelState.Feeds, feed)
			}
		}
		state.Channels[ch.SlackChannel] = channelState
	}
}

func processFeeds(cfg config.Config, state *st.State, slackClient *slack.Client) (int, int) {
	totalFeeds := 0
	totalNewPosts := 0

	for _, ch := range cfg.Channels {
		channel := ch.SlackChannel
		chState := state.Channels[channel]
		log.Printf("Processing channel: %s", channel)

		for _, feedURL := range ch.Feeds {
			totalFeeds++
			log.Printf("Checking feed: %s", feedURL)
			lastUpdated := chState.Feeds[feedURL].LastUpdated
			log.Printf("Last updated: %s", lastUpdated.Format(time.RFC3339))

			items, newLastUpdated, err := rss.FetchFeed(feedURL, lastUpdated)
			if err != nil {
				log.Printf("Error fetching feed %s: %v", feedURL, err)
				continue
			}

			log.Printf("Found %d new items in feed %s", len(items), feedURL)
			totalNewPosts += len(items)

			for _, item := range items {
				log.Printf("Posting new item to #%s: %s", channel, item.Title)
				if err := slackClient.PostMessage("#"+channel, rss.FormatItem(item)); err != nil {
					log.Printf("Error posting to Slack: %v", err)
				} else {
					log.Printf("Successfully posted to #%s", channel)
				}
			}

			if !newLastUpdated.Equal(lastUpdated) {
				log.Printf("Updating last updated time for %s to %s", feedURL, newLastUpdated.Format(time.RFC3339))
				chState.Feeds[feedURL] = st.FeedState{LastUpdated: newLastUpdated}
				state.Channels[channel] = chState
			}
		}
	}

	return totalFeeds, totalNewPosts
}
