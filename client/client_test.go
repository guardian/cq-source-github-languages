package client

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/rs/zerolog"
)

func TestNew(t *testing.T) {
	logger := zerolog.New(os.Stdout)
	ctx := context.Background()

	tests := []struct {
		name    string
		spec    *Spec
		wantErr bool
		errMsg  string
	}{
		{
			name:    "missing org",
			spec:    &Spec{},
			wantErr: true,
			errMsg:  "organization is required",
		},
		{
			name: "missing app_id",
			spec: &Spec{
				Org: "test-org",
			},
			wantErr: true,
			errMsg:  "github app id is required",
		},
		{
			name: "invalid app_id",
			spec: &Spec{
				Org:   "test-org",
				AppID: "not-a-number",
			},
			wantErr: true,
			errMsg:  "failed to parse app_id",
		},
		{
			name: "missing installation_id",
			spec: &Spec{
				Org:   "test-org",
				AppID: "12345",
			},
			wantErr: true,
			errMsg:  "github app installation id is required",
		},
		{
			name: "invalid installation_id",
			spec: &Spec{
				Org:            "test-org",
				AppID:          "12345",
				InstallationID: "not-a-number",
			},
			wantErr: true,
			errMsg:  "failed to parse installation_id",
		},
		{
			name: "missing private key",
			spec: &Spec{
				Org:            "test-org",
				AppID:          "12345",
				InstallationID: "67890",
			},
			wantErr: true,
			errMsg:  "github app private key is required",
		},
		{
			name: "invalid private key format",
			spec: &Spec{
				Org:            "test-org",
				AppID:          "12345",
				InstallationID: "67890",
				PrivateKey:     "invalid-key-content",
			},
			wantErr: true,
			errMsg:  "private key must be in PEM format",
		},
		{
			name: "valid config with private_key",
			spec: &Spec{
				Org:            "test-org",
				AppID:          "12345",
				InstallationID: "67890",
				PrivateKey:     "-----BEGIN RSA PRIVATE KEY-----\ntest-content\n-----END RSA PRIVATE KEY-----",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := New(ctx, logger, tt.spec)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("New() expected error but got none")
				} else if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("New() error = %v, expected to contain %v", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("New() unexpected error = %v", err)
				}
				if client.ID() != "github-languages" {
					t.Errorf("Client.ID() = %v, want %v", client.ID(), "github-languages")
				}
				if client.Org() != tt.spec.Org {
					t.Errorf("Client.Org() = %v, want %v", client.Org(), tt.spec.Org)
				}
			}
		})
	}
}

func TestNewWithPrivateKeyPath(t *testing.T) {
	logger := zerolog.New(os.Stdout)
	ctx := context.Background()

	// Create temporary key file
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "test-key.pem")
	keyContent := "-----BEGIN RSA PRIVATE KEY-----\ntest-content\n-----END RSA PRIVATE KEY-----"
	
	err := os.WriteFile(keyPath, []byte(keyContent), 0600)
	if err != nil {
		t.Fatalf("Failed to create test key file: %v", err)
	}

	spec := &Spec{
		Org:            "test-org",
		AppID:          "12345",
		InstallationID: "67890",
		PrivateKeyPath: keyPath,
	}

	client, err := New(ctx, logger, spec)
	if err != nil {
		t.Errorf("New() unexpected error = %v", err)
	}

	if client.PrivateKey != keyContent {
		t.Errorf("Private key not loaded correctly from file")
	}
}

func TestNewWithInvalidKeyPath(t *testing.T) {
	logger := zerolog.New(os.Stdout)
	ctx := context.Background()

	spec := &Spec{
		Org:            "test-org",
		AppID:          "12345",
		InstallationID: "67890",
		PrivateKeyPath: "/nonexistent/path/key.pem",
	}

	_, err := New(ctx, logger, spec)
	if err == nil {
		t.Errorf("New() expected error for invalid key path but got none")
	}
}

// Helper function for substring checking
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > len(substr) && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
