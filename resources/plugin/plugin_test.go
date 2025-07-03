package plugin

import (
	"context"
	"testing"

	"github.com/cloudquery/plugin-sdk/v4/plugin"
	"github.com/rs/zerolog"
)

func TestPlugin(t *testing.T) {
	p := Plugin()
	if p == nil {
		t.Fatal("Plugin() returned nil")
	}

	// Verify basic metadata is populated
	meta := p.Meta()
	if meta.Name == "" {
		t.Error("Plugin name should not be empty")
	}

	// Log the actual name for debugging
	t.Logf("Plugin name: %s", meta.Name)
}

func TestConfigureWithNoConnection(t *testing.T) {
	logger := zerolog.New(zerolog.NewTestWriter(t))

	client, err := Configure(context.Background(), logger, []byte("{}"), plugin.NewClientOptions{
		NoConnection: true,
	})

	if err != nil {
		t.Errorf("Configure() with NoConnection unexpected error = %v", err)
	}

	if client == nil {
		t.Error("Configure() returned nil client")
	}
}

func TestConfigureWithInvalidSpec(t *testing.T) {
	logger := zerolog.New(zerolog.NewTestWriter(t))

	_, err := Configure(context.Background(), logger, []byte("invalid-json"), plugin.NewClientOptions{})

	if err == nil {
		t.Error("Configure() expected error for invalid JSON but got none")
	}
}
