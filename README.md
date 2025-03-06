# Slack RSS Feed Manager

Manage Slack RSS feed subscriptions with a [config file](/config.yaml).

## How it works

- Add feed subscriptions per channel in the [config file](/config.yaml).
- A [GitHub action](/.github/workflows/rss-feed-check.yml) runs every hours sending Slack messages when new RSS items found.
