package state

import (
	"encoding/json"
	"os"
	"time"
)

type State struct {
	Channels map[string]ChannelState
}

type ChannelState struct {
	Feeds map[string]FeedState
}

type FeedState struct {
	LastUpdated time.Time
}

func LoadState(filePath string) (State, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return State{Channels: make(map[string]ChannelState)}, nil
	}
	var state State
	err = json.Unmarshal(data, &state)
	if state.Channels == nil {
		state.Channels = make(map[string]ChannelState)
	}
	return state, err
}

func (s *State) Save(filePath string) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0644)
}
