package slack

import (
	"errors"

	"github.com/slack-go/slack"
)

type Client struct {
	api *slack.Client
}

func NewClient(token string) *Client {
	return &Client{api: slack.New(token)}
}

func (c *Client) PostMessage(channel, text string) error {
	if channel == "" {
		return errors.New("channel cannot be empty")
	}
	if text == "" {
		return errors.New("message cannot be empty")
	}

	_, _, err := c.api.PostMessage(channel, slack.MsgOptionText(text, false))
	return err
}
