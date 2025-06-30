package github

import (
	"context"
	"testing"
)

func TestLanguagesStruct(t *testing.T) {
	langs := &Languages{
		FullName:  "owner/repo",
		Name:      "repo",
		Languages: []string{"Go", "JavaScript", "Python"},
	}

	if langs.FullName != "owner/repo" {
		t.Errorf("FullName = %v, want %v", langs.FullName, "owner/repo")
	}
	if langs.Name != "repo" {
		t.Errorf("Name = %v, want %v", langs.Name, "repo")
	}
	if len(langs.Languages) != 3 {
		t.Errorf("Languages length = %v, want %v", len(langs.Languages), 3)
	}
}

func TestNewGitHubAppClientWithInvalidKey(t *testing.T) {
	ctx := context.Background()
	
	tests := []struct {
		name           string
		appID          int64
		installationID int64
		privateKey     []byte
		wantErr        bool
	}{
		{
			name:           "empty private key",
			appID:          12345,
			installationID: 67890,
			privateKey:     []byte(""),
			wantErr:        true,
		},
		{
			name:           "invalid private key format",
			appID:          12345,
			installationID: 67890,
			privateKey:     []byte("not-a-valid-key"),
			wantErr:        true,
		},
		{
			name:           "short private key",
			appID:          12345,
			installationID: 67890,
			privateKey:     []byte("short"),
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewGitHubAppClient(ctx, tt.appID, tt.installationID, tt.privateKey)
			if tt.wantErr && err == nil {
				t.Errorf("NewGitHubAppClient() expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("NewGitHubAppClient() unexpected error = %v", err)
			}
		})
	}
}
