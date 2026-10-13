package github

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

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
		if err := json.NewEncoder(w).Encode(languages); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
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
		if _, err := w.Write([]byte(`{"message": "Not Found"}`)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
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

func TestClient_FetchRepositories(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/orgs/testorg/repos"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		repos := []map[string]any{
			{"name": "repo1", "archived": false},
			{"name": "repo2", "archived": true},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(repos); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	githubClient := github.NewClient(&http.Client{})
	githubClient.BaseURL, _ = url.Parse(server.URL + "/")

	client := &Client{
		GitHubClient: githubClient,
	}

	notArchived := func(repos []*github.Repository) []*github.Repository {
		var out []*github.Repository
		for _, r := range repos {
			if !r.GetArchived() {
				out = append(out, r)
			}
		}
		return out
	}

	result, err := client.FetchRepositories(context.Background(), "testorg", notArchived)
	if err != nil {
		t.Fatalf("FetchRepositories() error = %v", err)
	}

	if len(result) != 1 || result[0].GetName() != "repo1" {
		t.Errorf("FetchRepositories() = %v, want only repo1", result)
	}
}

func TestClient_FetchRepositoriesError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		if _, err := w.Write([]byte(`{"message": "Not Found"}`)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	githubClient := github.NewClient(&http.Client{})
	githubClient.BaseURL, _ = url.Parse(server.URL + "/")

	client := &Client{
		GitHubClient: githubClient,
	}

	_, err := client.FetchRepositories(context.Background(), "nonexistent", func(r []*github.Repository) []*github.Repository { return r })
	if err == nil {
		t.Error("FetchRepositories() expected error but got none")
	}
}

func TestClient_FetchRepositoriesPagination(t *testing.T) {
	var serverURL string
	pagesRequested := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pagesRequested++

		var repos []map[string]any
		switch r.URL.Query().Get("page") {
		case "":
			w.Header().Set("Link", `<`+serverURL+`/orgs/testorg/repos?page=2>; rel="next"`)
			repos = []map[string]any{{"name": "repo1"}}
		case "2":
			repos = []map[string]any{{"name": "repo2"}}
		default:
			t.Errorf("unexpected page %q", r.URL.Query().Get("page"))
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(repos); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()
	serverURL = server.URL

	githubClient := github.NewClient(&http.Client{})
	githubClient.BaseURL, _ = url.Parse(server.URL + "/")

	client := &Client{
		GitHubClient: githubClient,
	}

	result, err := client.FetchRepositories(context.Background(), "testorg", func(r []*github.Repository) []*github.Repository { return r })
	if err != nil {
		t.Fatalf("FetchRepositories() error = %v", err)
	}

	if pagesRequested != 2 {
		t.Errorf("pages requested = %d, want 2", pagesRequested)
	}

	if len(result) != 2 || result[0].GetName() != "repo1" || result[1].GetName() != "repo2" {
		t.Errorf("FetchRepositories() = %v, want repo1 and repo2", result)
	}
}

func newRetryTestClient(t *testing.T, failures int, okBody string) (*Client, *int) {
	t.Helper()
	oldDelay := retryDelay
	retryDelay = time.Millisecond
	t.Cleanup(func() { retryDelay = oldDelay })

	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		if calls <= failures {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"message": "Server Error"}`))
			return
		}
		_, _ = w.Write([]byte(okBody))
	}))
	t.Cleanup(server.Close)

	githubClient := github.NewClient(&http.Client{})
	githubClient.BaseURL, _ = url.Parse(server.URL + "/")
	return &Client{GitHubClient: githubClient}, &calls
}

func TestClient_GetLanguagesRetriesOnServerError(t *testing.T) {
	client, calls := newRetryTestClient(t, 2, `{"Go": 1}`)

	result, err := client.GetLanguages(context.Background(), "testowner", "testrepo")
	if err != nil {
		t.Fatalf("GetLanguages() error = %v", err)
	}
	if *calls != 3 {
		t.Errorf("calls = %d, want 3", *calls)
	}
	if len(result.Languages) != 1 || result.Languages[0] != "Go" {
		t.Errorf("Languages = %v, want [Go]", result.Languages)
	}
}

func TestClient_GetLanguagesGivesUpAfterRetries(t *testing.T) {
	client, calls := newRetryTestClient(t, 100, `{}`)

	_, err := client.GetLanguages(context.Background(), "testowner", "testrepo")
	if err == nil {
		t.Error("GetLanguages() expected error but got none")
	}
	if *calls != 3 {
		t.Errorf("calls = %d, want 3", *calls)
	}
}

func TestClient_FetchRepositoriesRetriesOnServerError(t *testing.T) {
	client, calls := newRetryTestClient(t, 2, `[{"name": "repo1"}]`)

	result, err := client.FetchRepositories(context.Background(), "testorg", func(r []*github.Repository) []*github.Repository { return r })
	if err != nil {
		t.Fatalf("FetchRepositories() error = %v", err)
	}
	if *calls != 3 {
		t.Errorf("calls = %d, want 3", *calls)
	}
	if len(result) != 1 || result[0].GetName() != "repo1" {
		t.Errorf("FetchRepositories() = %v, want repo1", result)
	}
}

func TestClient_FetchRepositoriesGivesUpAfterRetries(t *testing.T) {
	client, calls := newRetryTestClient(t, 100, `[]`)

	_, err := client.FetchRepositories(context.Background(), "testorg", func(r []*github.Repository) []*github.Repository { return r })
	if err == nil {
		t.Error("FetchRepositories() expected error but got none")
	}
	if *calls != 3 {
		t.Errorf("calls = %d, want 3", *calls)
	}
}

func TestWithRetry_StopsWhenContextCancelled(t *testing.T) {
	oldDelay := retryDelay
	retryDelay = time.Minute
	t.Cleanup(func() { retryDelay = oldDelay })

	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	_, _, err := withRetry(ctx, func() (int, *github.Response, error) {
		calls++
		cancel()
		return 0, &github.Response{Response: &http.Response{StatusCode: http.StatusInternalServerError}}, errors.New("boom")
	})

	if !errors.Is(err, context.Canceled) {
		t.Errorf("error = %v, want context.Canceled", err)
	}
	if calls != 1 {
		t.Errorf("calls = %d, want 1", calls)
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
