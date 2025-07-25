package main

import (
	"log"
	"os"
	"sort"
	"time"

	"slack-rss-feed-manager/config"
	"slack-rss-feed-manager/rss"
	"slack-rss-feed-manager/slack"
	st "slack-rss-feed-manager/state"
)

type SlackClient interface {
	PostMessage(channel, text string) error
}

type RSSClient interface {
	FetchFeed(url string, lastUpdated time.Time) ([]rss.FeedItem, time.Time, error)
}

type defaultRSSClient struct{}

func (c *defaultRSSClient) FetchFeed(url string, lastUpdated time.Time) ([]rss.FeedItem, time.Time, error) {
	return rss.FetchFeed(url, lastUpdated)
}

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
	rssClient := &defaultRSSClient{}
	updateSubscriptions(cfg, &currentState, rssClient)
	log.Printf("Processing feeds...")
	feedsProcessed, postsFound := processFeeds(cfg, &currentState, slackClient, rssClient)

	// Save updated state
	log.Printf("Saving updated state to %s", stateFile)
	if err := currentState.Save(stateFile); err != nil {
		log.Fatalf("Failed to save state: %v", err)
	}

	duration := time.Since(startTime)
	log.Printf("RSS Feed Manager completed successfully in %v", duration)
	log.Printf("Summary: Processed %d feeds, found %d new posts", feedsProcessed, postsFound)
}

func updateSubscriptions(cfg config.Config, state *st.State, rssClient RSSClient) {
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
				
				// For new feeds, fetch the latest post and set LastUpdated to 1 hour before it
				// This ensures the most recent post will be picked up in the next processing cycle
				items, _, err := rssClient.FetchFeed(feed, time.Time{}) // Use zero time to get all items
				if err != nil {
					log.Printf("Warning: Failed to fetch new feed %s for initial setup: %v", feed, err)
					// Fallback to 24 hours ago if we can't fetch the feed
					channelState.Feeds[feed] = st.FeedState{LastUpdated: time.Now().Add(-24 * time.Hour)}
				} else if len(items) > 0 {
					// Set LastUpdated to 1 hour before the most recent post
					mostRecent := items[0].Published
					for _, item := range items {
						if item.Published.After(mostRecent) {
							mostRecent = item.Published
						}
					}
					channelState.Feeds[feed] = st.FeedState{LastUpdated: mostRecent.Add(-1 * time.Hour)}
					log.Printf("New feed %s: most recent post at %s, set LastUpdated to %s", 
						feed, mostRecent.Format(time.RFC3339), mostRecent.Add(-1*time.Hour).Format(time.RFC3339))
				} else {
					// No items in feed, set to current time
					channelState.Feeds[feed] = st.FeedState{LastUpdated: time.Now()}
					log.Printf("New feed %s has no items, set LastUpdated to now", feed)
				}
			}
		}
		// Remove feeds not in config
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

func processFeeds(cfg config.Config, state *st.State, slackClient SlackClient, rssClient RSSClient) (int, int) {
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

			items, newLastUpdated, err := rssClient.FetchFeed(feedURL, lastUpdated)
			if err != nil {
				log.Printf("Error fetching feed %s: %v", feedURL, err)
				continue
			}

			log.Printf("Found %d new items in feed %s", len(items), feedURL)
			totalNewPosts += len(items)

			sort.Slice(items, func(i, j int) bool {
				return items[i].Published.Before(items[j].Published)
			})

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
