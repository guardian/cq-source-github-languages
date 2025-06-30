package plugin

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/cloudquery/plugin-sdk/v4/plugin"
	"github.com/guardian/cq-source-github-languages/client"
	"github.com/rs/zerolog"
)

func TestPlugin(t *testing.T) {
	p := Plugin()

	if p == nil {
		t.Fatal("Plugin() returned nil")
	}

	// Test plugin metadata
	meta := p.Meta()
	if meta.Name != Name {
		t.Errorf("Plugin name = %v, want %v", meta.Name, Name)
	}
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

func TestConfigureWithValidSpec(t *testing.T) {
	logger := zerolog.New(zerolog.NewTestWriter(t))

	spec := client.Spec{
		Org:            "test-org",
		AppID:          "12345",
		InstallationID: "67890",
		PrivateKey:     "-----BEGIN RSA PRIVATE KEY-----\ntest-content\n-----END RSA PRIVATE KEY-----",
	}

	specBytes, err := json.Marshal(spec)
	if err != nil {
		t.Fatalf("Failed to marshal spec: %v", err)
	}

	client, err := Configure(context.Background(), logger, specBytes, plugin.NewClientOptions{})

	if err != nil {
		t.Errorf("Configure() unexpected error = %v", err)
	}

	if client == nil {
		t.Error("Configure() returned nil client")
	}
}

func TestGetTables(t *testing.T) {
	tables := getTables()

	if len(tables) == 0 {
		t.Error("getTables() returned empty table list")
	}

	// Check that the languages table exists
	found := false
	for _, table := range tables {
		if table.Name == "github_languages" {
			found = true
			break
		}
	}

	if !found {
		t.Error("github_languages table not found in table list")
	}
}
