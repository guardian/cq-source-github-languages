package client

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/rs/zerolog"
)

const (
	testPEMKey = "-----BEGIN RSA PRIVATE KEY-----\ntest-content\n-----END RSA PRIVATE KEY-----"
	testOrg    = "test-org"
	testAppID  = "12345"
	testInstID = "67890"
)

func testLogger(t *testing.T) zerolog.Logger {
	return zerolog.New(zerolog.NewTestWriter(t))
}

func TestNew(t *testing.T) {
	logger := testLogger(t)
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
				Org: testOrg,
			},
			wantErr: true,
			errMsg:  "github app id is required",
		},
		{
			name: "invalid app_id",
			spec: &Spec{
				Org:   testOrg,
				AppID: "not-a-number",
			},
			wantErr: true,
			errMsg:  "failed to parse app_id",
		},
		{
			name: "missing installation_id",
			spec: &Spec{
				Org:   testOrg,
				AppID: testAppID,
			},
			wantErr: true,
			errMsg:  "github app installation id is required",
		},
		{
			name: "invalid installation_id",
			spec: &Spec{
				Org:            testOrg,
				AppID:          testAppID,
				InstallationID: "not-a-number",
			},
			wantErr: true,
			errMsg:  "failed to parse installation_id",
		},
		{
			name: "missing private key",
			spec: &Spec{
				Org:            testOrg,
				AppID:          testAppID,
				InstallationID: testInstID,
			},
			wantErr: true,
			errMsg:  "github app private key is required",
		},
		{
			name: "invalid private key format",
			spec: &Spec{
				Org:            testOrg,
				AppID:          testAppID,
				InstallationID: testInstID,
				PrivateKey:     "invalid-key-content",
			},
			wantErr: true,
			errMsg:  "private key must be in PEM format",
		},
		{
			name: "valid config with private_key",
			spec: &Spec{
				Org:            testOrg,
				AppID:          testAppID,
				InstallationID: testInstID,
				PrivateKey:     testPEMKey,
			},
			wantErr: false,
		},
		{
			name: "whitespace in app_id",
			spec: &Spec{
				Org:            testOrg,
				AppID:          "  12345  ",
				InstallationID: testInstID,
				PrivateKey:     testPEMKey,
			},
			wantErr: false, // Should be trimmed and work
		},
		{
			name: "file interpolation syntax warning",
			spec: &Spec{
				Org:            testOrg,
				AppID:          "${file:app-id.txt}",
				InstallationID: testInstID,
				PrivateKey:     testPEMKey,
			},
			wantErr: true, // Should fail parsing but log warning
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := New(ctx, logger, tt.spec)

			if tt.wantErr {
				if err == nil {
					t.Errorf("New() expected error but got none")
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
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
	logger := testLogger(t)
	ctx := context.Background()

	// Create temporary key file
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "test-key.pem")
	keyContent := testPEMKey

	err := os.WriteFile(keyPath, []byte(keyContent), 0600)
	if err != nil {
		t.Fatalf("Failed to create test key file: %v", err)
	}

	spec := &Spec{
		Org:            testOrg,
		AppID:          testAppID,
		InstallationID: testInstID,
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
	logger := testLogger(t)
	ctx := context.Background()

	spec := &Spec{
		Org:            testOrg,
		AppID:          testAppID,
		InstallationID: testInstID,
		PrivateKeyPath: "/nonexistent/path/key.pem",
	}

	_, err := New(ctx, logger, spec)
	if err == nil {
		t.Errorf("New() expected error for invalid key path but got none")
	}
}

func TestPrivateKeyPrecedence(t *testing.T) {
	logger := testLogger(t)

	// Create temp file
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "key.pem")
	fileContent := "-----BEGIN RSA PRIVATE KEY-----\nfile-content\n-----END RSA PRIVATE KEY-----"

	if err := os.WriteFile(keyPath, []byte(fileContent), 0600); err != nil {
		t.Fatalf("Failed to create key file: %v", err)
	}

	// Test that PrivateKeyPath takes precedence over PrivateKey
	spec := &Spec{
		Org:            testOrg,
		AppID:          testAppID,
		InstallationID: testInstID,
		PrivateKey:     testPEMKey,
		PrivateKeyPath: keyPath,
	}

	client, err := New(context.Background(), logger, spec)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if client.PrivateKey != fileContent {
		t.Errorf("Expected file content to take precedence, got %v", client.PrivateKey)
	}
}
func TestEdgeCases(t *testing.T) {
	logger := testLogger(t)
	ctx := context.Background()

	tests := []struct {
		name    string
		spec    *Spec
		wantErr bool
		errMsg  string
	}{
		{
			name: "whitespace in installation_id",
			spec: &Spec{
				Org:            testOrg,
				AppID:          testAppID,
				InstallationID: "  67890  ",
				PrivateKey:     testPEMKey,
			},
			wantErr: false, // Should be trimmed and work
		},
		{
			name: "file interpolation syntax in installation_id",
			spec: &Spec{
				Org:            testOrg,
				AppID:          testAppID,
				InstallationID: "${file:install-id.txt}",
				PrivateKey:     testPEMKey,
			},
			wantErr: true, // Should fail parsing but log warning
		},
		{
			name: "private key with only BEGIN marker",
			spec: &Spec{
				Org:            testOrg,
				AppID:          testAppID,
				InstallationID: testInstID,
				PrivateKey:     "-----BEGIN RSA PRIVATE KEY-----\ntest-content",
			},
			wantErr: true,
			errMsg:  "private key must be in PEM format",
		},
		{
			name: "private key with only END marker",
			spec: &Spec{
				Org:            testOrg,
				AppID:          testAppID,
				InstallationID: testInstID,
				PrivateKey:     "test-content\n-----END RSA PRIVATE KEY-----",
			},
			wantErr: true,
			errMsg:  "private key must be in PEM format",
		},
		{
			name: "private key with whitespace around PEM",
			spec: &Spec{
				Org:            testOrg,
				AppID:          testAppID,
				InstallationID: testInstID,
				PrivateKey:     "  " + testPEMKey + "  ",
			},
			wantErr: false, // Should be trimmed and work
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := New(ctx, logger, tt.spec)

			if tt.wantErr {
				if err == nil {
					t.Errorf("New() expected error but got none")
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
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

func TestNewWithEmptyPrivateKeyFile(t *testing.T) {
	logger := testLogger(t)
	ctx := context.Background()

	// Create temporary empty key file
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "empty-key.pem")

	err := os.WriteFile(keyPath, []byte(""), 0600)
	if err != nil {
		t.Fatalf("Failed to create empty key file: %v", err)
	}

	spec := &Spec{
		Org:            testOrg,
		AppID:          testAppID,
		InstallationID: testInstID,
		PrivateKeyPath: keyPath,
	}

	_, err = New(ctx, logger, spec)
	if err == nil {
		t.Errorf("New() expected error for empty key file but got none")
	}
	if !strings.Contains(err.Error(), "github app private key is required") {
		t.Errorf("New() error = %v, expected to contain 'github app private key is required'", err)
	}
}

func TestNewWithWhitespaceOnlyPrivateKeyFile(t *testing.T) {
	logger := testLogger(t)
	ctx := context.Background()

	// Create temporary key file with only whitespace
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "whitespace-key.pem")

	err := os.WriteFile(keyPath, []byte("   \n\t  \n  "), 0600)
	if err != nil {
		t.Fatalf("Failed to create whitespace key file: %v", err)
	}

	spec := &Spec{
		Org:            testOrg,
		AppID:          testAppID,
		InstallationID: testInstID,
		PrivateKeyPath: keyPath,
	}

	_, err = New(ctx, logger, spec)
	if err == nil {
		t.Errorf("New() expected error for whitespace-only key file but got none")
	}
	if !strings.Contains(err.Error(), "github app private key is required") {
		t.Errorf("New() error = %v, expected to contain 'github app private key is required'", err)
	}
}

func TestClientMethods(t *testing.T) {
	logger := testLogger(t)
	ctx := context.Background()

	spec := &Spec{
		Org:            testOrg,
		AppID:          testAppID,
		InstallationID: testInstID,
		PrivateKey:     testPEMKey,
	}

	client, err := New(ctx, logger, spec)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test ID method
	if got := client.ID(); got != "github-languages" {
		t.Errorf("Client.ID() = %v, want %v", got, "github-languages")
	}

	// Test Logger method
	if logger := client.Logger(); logger == nil {
		t.Errorf("Client.Logger() returned nil")
	}

	// Test Org method
	if got := client.Org(); got != testOrg {
		t.Errorf("Client.Org() = %v, want %v", got, testOrg)
	}

	// Test that parsed values are correct
	expectedAppID, _ := strconv.ParseInt(testAppID, 10, 64)
	if client.AppID != expectedAppID {
		t.Errorf("Client.AppID = %v, want %v", client.AppID, expectedAppID)
	}

	expectedInstID, _ := strconv.ParseInt(testInstID, 10, 64)
	if client.InstallationID != expectedInstID {
		t.Errorf("Client.InstallationID = %v, want %v", client.InstallationID, expectedInstID)
	}

	if client.PrivateKey != testPEMKey {
		t.Errorf("Client.PrivateKey = %v, want %v", client.PrivateKey, testPEMKey)
	}
}
