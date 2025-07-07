package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/google/go-github/v57/github"
)

const (
	dummyRSAKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA1234567890abcdef...
-----END RSA PRIVATE KEY-----`
	testAppID          = int64(12345) // int64 - matches API layer
	testInstallationID = int64(67890) // int64 - matches API layer
)

func TestClient_GetLanguages(t *testing.T) {
	// Mock GitHub API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check the request path
		expectedPath := "/repos/testowner/testrepo/languages"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		// Mock response - languages endpoint returns a map of language -> bytes
		languages := map[string]int{
			"Go":         12345,
			"JavaScript": 6789,
			"Python":     3456,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(languages)
	}))
	defer server.Close()

	// Create a GitHub client pointing to our test server
	githubClient := github.NewClient(&http.Client{})
	githubClient.BaseURL, _ = url.Parse(server.URL + "/")

	client := &Client{
		GitHubClient: githubClient,
	}

	// Test GetLanguages
	ctx := context.Background()
	result, err := client.GetLanguages(ctx, "testowner", "testrepo")

	if err != nil {
		t.Fatalf("GetLanguages() error = %v", err)
	}

	// Verify the result
	expectedFullName := "testowner/testrepo"
	if result.FullName != expectedFullName {
		t.Errorf("FullName = %v, want %v", result.FullName, expectedFullName)
	}

	expectedName := "testrepo"
	if result.Name != expectedName {
		t.Errorf("Name = %v, want %v", result.Name, expectedName)
	}

	expectedLanguages := []string{"Go", "JavaScript", "Python"}
	if len(result.Languages) != len(expectedLanguages) {
		t.Errorf("Languages length = %d, want %d", len(result.Languages), len(expectedLanguages))
	}

	// Check that all expected languages are present (order might vary due to map iteration)
	languageMap := make(map[string]bool)
	for _, lang := range result.Languages {
		languageMap[lang] = true
	}

	for _, expectedLang := range expectedLanguages {
		if !languageMap[expectedLang] {
			t.Errorf("Expected language %s not found in result", expectedLang)
		}
	}
}

func TestClient_GetLanguagesError(t *testing.T) {
	// Mock GitHub API server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message": "Not Found"}`))
	}))
	defer server.Close()

	githubClient := github.NewClient(&http.Client{})
	githubClient.BaseURL, _ = url.Parse(server.URL + "/")

	client := &Client{
		GitHubClient: githubClient,
	}

	ctx := context.Background()
	_, err := client.GetLanguages(ctx, "nonexistent", "repo")

	if err == nil {
		t.Error("GetLanguages() expected error but got none")
	}
}

func TestNewGitHubAppClient(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		appID          int64
		installationID int64
		privateKey     []byte
		wantErr        bool
		expectedErrMsg string
	}{
		{
			name:           "empty private key",
			appID:          testAppID,
			installationID: testInstallationID,
			privateKey:     []byte(""),
			wantErr:        true,
			expectedErrMsg: "appears too short",
		},
		{
			name:           "invalid PEM format",
			appID:          testAppID,
			installationID: testInstallationID,
			privateKey:     []byte("not-pem-data"),
			wantErr:        true,
			expectedErrMsg: "appears too short",
		},
		{
			name:           "empty PEM block",
			appID:          testAppID,
			installationID: testInstallationID,
			privateKey:     []byte("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"),
			wantErr:        true,
			expectedErrMsg: "appears too short",
		},
		{
			name:           "short key",
			appID:          testAppID,
			installationID: testInstallationID,
			privateKey:     []byte("short"),
			wantErr:        true,
			expectedErrMsg: "appears too short",
		},
		{
			name:           "dummy RSA key (valid PEM, invalid RSA)",
			appID:          testAppID,
			installationID: testInstallationID,
			privateKey:     []byte(dummyRSAKey),
			wantErr:        true,
			expectedErrMsg: "appears too short",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewGitHubAppClient(ctx, tt.appID, tt.installationID, tt.privateKey)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				if !strings.Contains(err.Error(), tt.expectedErrMsg) {
					t.Errorf("error = %v, expected to contain %v", err, tt.expectedErrMsg)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error = %v", err)
				}
			}
		})
	}
}
