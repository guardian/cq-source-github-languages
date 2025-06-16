package github

import (
	"context"
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
func NewGitHubAppClient(appID, installationID int64, privateKeyPEM []byte) (*Client, error) {
	// Create a new transport using the GitHub App authentication
	itr, err := newGitHubAppTransport(appID, installationID, privateKeyPEM)
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
func newGitHubAppTransport(appID, installationID int64, privateKeyPEM []byte) (http.RoundTripper, error) {
	// Parse the private key
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		// Provide more specific error for key parsing issues
		return nil, fmt.Errorf("failed to parse private key: %w - ensure key is a valid PEM-encoded RSA private key", err)
	}

	// Create JWT token for GitHub App
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iat": jwt.NewNumericDate(jwt.TimeFunc().Add(-30 * time.Second)),
		"exp": jwt.NewNumericDate(jwt.TimeFunc().Add(5 * time.Minute)),
		"iss": appID,
	})

	signedToken, err := token.SignedString(privateKey)
	if err != nil {
		return nil, err
	}

	// Create an authenticated client for getting an installation token
	client := github.NewClient(&http.Client{
		Transport: &oauth2.Transport{
			Source: oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: signedToken},
			),
		},
	})

	// Get installation token
	installToken, _, err := client.Apps.CreateInstallationToken(
		context.Background(),
		installationID,
		&github.InstallationTokenOptions{},
	)
	if err != nil {
		return nil, err
	}

	// Return transport that uses the installation token
	return &oauth2.Transport{
		Source: oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: *installToken.Token},
		),
	}, nil
}

func (c *Client) GetLanguages(owner string, name string) (*Languages, error) {
	langs, _, err := c.GitHubClient.Repositories.ListLanguages(context.Background(), owner, name)
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
