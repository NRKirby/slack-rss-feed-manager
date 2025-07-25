# Slack RSS Feed Manager

## Project Overview

This is a Go-based RSS feed manager that monitors RSS feeds and posts new items to Slack channels. The application is designed to run as a scheduled job (GitHub Actions) to automatically check for new RSS feed items and notify Slack channels when new content is found.

## Architecture

The application is structured as a modular Go project with distinct packages for different responsibilities:

- **cmd/main.go**: Main application entry point and orchestration logic
- **config/**: Configuration loading and validation
- **rss/**: RSS feed parsing and processing
- **slack/**: Slack API integration for posting messages
- **state/**: State persistence to track last-seen feed items

## Key Files & Structure

```
/
├── CLAUDE.md              # This file - project documentation for Claude
├── README.md              # Basic project documentation
├── go.mod                 # Go module dependencies
├── go.sum                 # Go module checksums
├── justfile               # Just command runner with run/test commands
├── config.yaml            # Feed configuration (channels and RSS URLs)
├── state.json             # Persistent state tracking last update times
├── cmd/
│   ├── main.go           # Main application logic and orchestration
│   └── main_test.go      # Tests for main application
├── config/
│   ├── config.go         # Configuration loading and validation
│   └── config_test.go    # Configuration tests
├── rss/
│   ├── rss.go           # RSS feed fetching and parsing
│   └── rss_test.go      # RSS package tests
├── slack/
│   ├── slack.go         # Slack API client and message posting
│   └── slack_test.go    # Slack package tests
├── state/
│   └── state.go         # State persistence for tracking feed updates
└── .github/workflows/
    └── rss-feed-check.yml # GitHub Actions workflow for automation
```

## Core Functionality

### Main Application Flow (cmd/main.go)

1. Load configuration from `config.yaml`
2. Load existing state from `state.json`
3. Update subscriptions based on config changes
4. Process each feed for each channel:
   - Fetch RSS feed using gofeed library
   - Filter items newer than last update time
   - Sort items by publication date
   - Post new items to corresponding Slack channel
   - Update state with new last-updated time
5. Save updated state back to `state.json`

### Configuration (config/)

- Loads YAML configuration from `config.yaml`
- Validates that channels and feeds are properly configured
- Structure: channels contain slack_channel name and array of feed URLs

### RSS Processing (rss/)

- Uses `github.com/mmcdole/gofeed` library for RSS parsing
- Filters items based on last-updated timestamp
- Returns structured FeedItem with title, link, published time, and feed title
- Formats messages for Slack posting

### Slack Integration (slack/)

- Uses `github.com/slack-go/slack` library
- Simple wrapper around Slack API for posting messages
- Requires SLACK_BOT_TOKEN environment variable

### State Management (state/)

- JSON-based persistence in `state.json`
- Tracks last-updated time per feed per channel
- Gracefully handles missing state file (creates new state)

## Dependencies

Key Go modules:

- `github.com/mmcdole/gofeed v1.3.0` - RSS/Atom feed parsing
- `github.com/slack-go/slack v0.16.0` - Slack API client
- `gopkg.in/yaml.v3 v3.0.1` - YAML configuration parsing

## Configuration

### config.yaml Format

```yaml
channels:
  - slack_channel: channel-name-without-hash
    feeds:
      - https://example.com/rss
      - https://another-example.com/feed.xml
```

### Environment Variables

- `SLACK_BOT_TOKEN`: Required Slack bot token for API access

## Automation

### GitHub Actions (.github/workflows/rss-feed-check.yml)

- Runs every hour between 7am-10pm UTC (cron: '0 7-22 \* \* \*')
- Also runs on push to main branch and manual dispatch
- Sets up Go 1.21, runs the application, commits state changes
- Requires `SLACK_BOT_TOKEN` as a GitHub secret

## Testing

- Each package has corresponding `*_test.go` files
- Run tests with: `just test` or `go test ./... -v`
- Main application uses interfaces (SlackClient, RSSClient) for testability

## Development Commands

Using Just (justfile):

- `just run` - Run the application locally
- `just test` - Run all tests

Using Go directly:

- `go run cmd/main.go` - Run the application
- `go test ./... -v` - Run tests with verbose output

## Current Configuration

The project is currently configured to monitor:

- Channel: `tech-blog-alerts`
- Feeds:
  - https://antirez.com/rss (Redis creator's blog)
  - https://deepmind.google/blog/rss.xml (DeepMind blog)
  - https://go.dev/blog/feed.atom (Go blog)
  - https://michael.stapelberg.ch/feed.xml (Go developer's blog)
  - https://tailscale.com/blog/index.xml (Tailscale blog)

## State Tracking

The application maintains state in `state.json` to avoid reposting old items. The state file tracks the last-updated timestamp for each feed in each channel, allowing the application to only post new items since the last run.

## Error Handling

- Graceful handling of missing configuration/state files
- Feed parsing errors are logged but don't stop processing other feeds
- Slack posting errors are logged but don't stop processing
- Application will exit with fatal error only for critical issues (missing token, invalid config)

## Message Format

RSS items are posted to Slack in the format:

```
New post from {FeedTitle}: {ItemTitle}
{ItemLink}
```
