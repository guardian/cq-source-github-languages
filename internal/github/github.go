package github

import (
	"context"
	"encoding/pem"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/go-github/v57/github"
	"golang.org/x/exp/maps"
	"golang.org/x/oauth2"
)

type Languages struct {
	FullName  string
	Name      string
	Languages []string
}

type Client struct {
	GitHubClient *github.Client
}

// NewGitHubAppClient creates a new GitHub client authenticated as a GitHub App installation
func NewGitHubAppClient(ctx context.Context, appID, installationID int64, privateKeyPEM []byte) (*Client, error) {
	// Create a new transport using the GitHub App authentication
	itr, err := newGitHubAppTransport(ctx, appID, installationID, privateKeyPEM)
	if err != nil {
		return nil, err
	}

	// Create a new client with the transport
	httpClient := &http.Client{Transport: itr}
	client := github.NewClient(httpClient)

	return &Client{
		GitHubClient: client,
	}, nil
}

// newGitHubAppTransport creates a new http.RoundTripper that authenticates as a GitHub App installation
func newGitHubAppTransport(ctx context.Context, appID, installationID int64, privateKeyPEM []byte) (http.RoundTripper, error) {
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		// Debug: Check if the key looks like a valid PEM key
		keyStr := string(privateKeyPEM)
		if len(keyStr) < 100 {
			return nil, fmt.Errorf("private key appears too short (%d chars) - ensure full key is provided", len(keyStr))
		}
		// Use encoding/pem to decode the PEM block
		block, _ := pem.Decode([]byte(keyStr))
		if block == nil {
			return nil, fmt.Errorf("private key is not a valid PEM format: %w", err)
		}
		return nil, fmt.Errorf("failed to parse RSA private key from PEM: %w - ensure key is RSA format and not EC/Ed25519", err)
	}

	// Create JWT token with proper timing and claims
	now := time.Now()
	claims := jwt.MapClaims{
		"iat": jwt.NewNumericDate(now.Add(-60 * time.Second)), // 1 minute ago to account for clock skew
		"exp": jwt.NewNumericDate(now.Add(10 * time.Minute)),  // 10 minutes from now (max allowed)
		"iss": appID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	signedToken, err := token.SignedString(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign JWT token: %w", err)
	}

	// Debug: Print token claims for debugging
	fmt.Printf("JWT Claims: iss=%d, iat=%v, exp=%v\n",
		appID,
		claims["iat"].(*jwt.NumericDate).Time,
		claims["exp"].(*jwt.NumericDate).Time)

	// Create an authenticated client for getting an installation token
	client := github.NewClient(&http.Client{
		Transport: &oauth2.Transport{
			Source: oauth2.StaticTokenSource(
				&oauth2.Token{
					AccessToken: signedToken,
					TokenType:   "Bearer",
				},
			),
		},
	})

	// Test the JWT token first by trying to list app installations
	fmt.Printf("Testing JWT token by listing app installations...\n")
	installations, _, err := client.Apps.ListInstallations(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("JWT token validation failed when listing installations: %w - check App ID (%d) and private key", err, appID)
	}
	fmt.Printf("JWT token valid - found %d installations\n", len(installations))

	// Get installation token with timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	fmt.Printf("Attempting to create installation token for installation ID: %d\n", installationID)

	installToken, resp, err := client.Apps.CreateInstallationToken(
		timeoutCtx,
		installationID,
		&github.InstallationTokenOptions{},
	)
	if err != nil {
		// Provide more detailed error information
		if resp != nil {
			return nil, fmt.Errorf("failed to create installation token (HTTP %d): %w - verify App ID (%d) and Installation ID (%d) are correct",
				resp.StatusCode, err, appID, installationID)
		}
		return nil, fmt.Errorf("failed to create installation token: %w - verify App ID (%d) and Installation ID (%d) are correct",
			err, appID, installationID)
	}

	if installToken == nil || installToken.Token == nil {
		return nil, fmt.Errorf("received nil installation token")
	}

	fmt.Printf("Successfully created installation token (expires: %v)\n", installToken.ExpiresAt)

	// Return transport that uses the installation token
	return &oauth2.Transport{
		Source: oauth2.StaticTokenSource(
			&oauth2.Token{
				AccessToken: *installToken.Token,
				TokenType:   "token",
			},
		),
	}, nil
}

func (c *Client) GetLanguages(ctx context.Context, owner string, name string) (*Languages, error) {
	langs, _, err := c.GitHubClient.Repositories.ListLanguages(ctx, owner, name)
	if err != nil {
		return nil, err
	}
	l := &Languages{
		FullName:  owner + "/" + name,
		Name:      name,
		Languages: maps.Keys(langs),
	}
	return l, nil

}
