package slack

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	t.Run("creates client with token", func(t *testing.T) {
		client := NewClient("test-token")
		if client == nil {
			t.Error("Expected non-nil client")
		}
	})
}

func TestPostMessage(t *testing.T) {
	tests := []struct {
		name        string
		channel     string
		message     string
		expectError bool
	}{
		{
			name:        "empty channel",
			channel:     "",
			message:     "Test message",
			expectError: true,
		},
		{
			name:        "empty message",
			channel:     "test-channel",
			message:     "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient("test-token") // Using real client for input validation
			err := client.PostMessage(tt.channel, tt.message)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}
