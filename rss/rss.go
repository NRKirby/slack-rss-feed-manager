package rss

import (
	"fmt"
	"time"

	"github.com/mmcdole/gofeed"
)

type FeedItem struct {
	Title     string
	Link      string
	Published time.Time
	FeedTitle string
}

func FetchFeed(url string, lastUpdated time.Time) ([]FeedItem, time.Time, error) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(url)
	if err != nil {
		return nil, lastUpdated, err
	}
	var items []FeedItem
	latest := lastUpdated
	for _, item := range feed.Items {
		pubTime := item.PublishedParsed
		if pubTime == nil {
			pubTime = item.UpdatedParsed
		}
		if pubTime != nil && pubTime.After(lastUpdated) {
			items = append(items, FeedItem{
				Title:     item.Title,
				Link:      item.Link,
				Published: *pubTime,
				FeedTitle: feed.Title,
			})
			if pubTime.After(latest) {
				latest = *pubTime
			}
		}
	}
	return items, latest, nil
}

func FormatItem(item FeedItem) string {
	return fmt.Sprintf("New post from %s: %s\n%s", item.FeedTitle, item.Title, item.Link)
}
