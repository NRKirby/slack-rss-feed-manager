# Slack RSS Feed Manager

Manage Slack RSS feed subscriptions with a [config file](/config.yaml).

## How it works

- Add feed subscriptions per channel in the [config file](/config.yaml).
- A [GitHub action](/.github/workflows/rss-feed-check.yml) runs every 4 hours and sends Slack messages for new RSS items found.
