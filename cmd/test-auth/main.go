package main

import (
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/go-github/v57/github"
	"github.com/guardian/cq-source-github-languages/client"
	githubinternal "github.com/guardian/cq-source-github-languages/internal/github"
	"github.com/rs/zerolog"
	"golang.org/x/oauth2"
)

// sanitizePrivateKey ensures the private key is properly formatted with newlines
// and handles base64 encoded keys
func sanitizePrivateKey(key string) ([]byte, error) {
	// Trim any extra whitespace
	key = strings.TrimSpace(key)

	// Check if the key is base64 encoded (doesn't start with PEM headers)
	if !strings.HasPrefix(key, "-----BEGIN") {
		// Try to decode as base64
		fmt.Println("Key doesn't start with PEM header, attempting base64 decode...")
		decoded, err := base64.StdEncoding.DecodeString(key)
		if err != nil {
			return nil, fmt.Errorf("key doesn't appear to be in PEM format and failed base64 decoding: %w", err)
		}

		// Check if decoded content looks like a PEM key
		decodedStr := string(decoded)
		if strings.Contains(decodedStr, "-----BEGIN") {
			fmt.Println("Successfully decoded base64 key to PEM format")
			key = decodedStr
		} else {
			return nil, fmt.Errorf("base64 decoded content doesn't appear to be a valid PEM key")
		}
	}

	// Replace literal "\n" with actual newlines if they exist
	key = strings.ReplaceAll(key, "\\n", "\n")

	// Validate the key format more thoroughly
	if !strings.Contains(key, "-----BEGIN") {
		return nil, fmt.Errorf("private key does not contain BEGIN marker")
	}
	if !strings.Contains(key, "-----END") {
		return nil, fmt.Errorf("private key does not contain END marker")
	}

	// Check for supported key types
	supportedTypes := []string{
		"-----BEGIN RSA PRIVATE KEY-----",
		"-----BEGIN PRIVATE KEY-----", // PKCS#8 format
	}

	isSupported := false
	keyType := "unknown"
	for _, keyTypeMarker := range supportedTypes {
		if strings.Contains(key, keyTypeMarker) {
			isSupported = true
			if strings.Contains(keyTypeMarker, "RSA") {
				keyType = "RSA (PKCS#1)"
			} else {
				keyType = "PKCS#8"
			}
			break
		}
	}

	if !isSupported {
		return nil, fmt.Errorf("unsupported private key format - must be RSA PKCS#1 or PKCS#8 format")
	}

	fmt.Printf("Detected private key format: %s\n", keyType)
	fmt.Printf("Private key length: %d characters\n", len(key))

	return []byte(key), nil
}

// createJWTAuthenticatedClient creates a GitHub client authenticated with JWT (for app-level operations)
func createJWTAuthenticatedClient(appID int64, privateKeyPEM []byte) (*github.Client, error) {
	// Parse the private key
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	// Create JWT token with proper claims
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iat": jwt.NewNumericDate(now.Add(-60 * time.Second)), // 1 minute ago to account for clock skew
		"exp": jwt.NewNumericDate(now.Add(10 * time.Minute)),  // 10 minutes from now (max allowed)
		"iss": appID,
	})

	// Sign the token
	signedToken, err := token.SignedString(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign JWT token: %w", err)
	}

	// Create client with JWT authentication
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

	return client, nil
}

func main() {
	// Define flags for command line arguments
	// NOTE: Command line flags take precedence over environment variables
	// If both are provided, the CLI flag value will be used
	appIDStr := flag.String("app-id", os.Getenv("GITHUB_APP_ID"), "GitHub App ID (or set GITHUB_APP_ID env)")
	installIDStr := flag.String("install-id", os.Getenv("GITHUB_INSTALLATION_ID"), "GitHub Installation ID (or set GITHUB_INSTALLATION_ID env)")
	keyPath := flag.String("key-path", os.Getenv("GITHUB_PRIVATE_KEY_PATH"), "Path to GitHub App private key file (or set GITHUB_PRIVATE_KEY_PATH env)")
	key := flag.String("key", os.Getenv("GITHUB_PRIVATE_KEY"), "GitHub App private key contents directly (or set GITHUB_PRIVATE_KEY env)")
	owner := flag.String("owner", "guardian", "GitHub repository owner/organization")
	repo := flag.String("repo", "cq-source-github-languages", "GitHub repository name")

	flag.Parse()

	// Show which values are being used for debugging
	fmt.Println("Configuration:")
	if *appIDStr != "" {
		fmt.Printf("App ID: %s (from %s)\n", *appIDStr, getValueSource("GITHUB_APP_ID", *appIDStr))
	}
	if *installIDStr != "" {
		fmt.Printf("Installation ID: %s (from %s)\n", *installIDStr, getValueSource("GITHUB_INSTALLATION_ID", *installIDStr))
	}
	if *keyPath != "" {
		fmt.Printf("Key Path: %s (from %s)\n", *keyPath, getValueSource("GITHUB_PRIVATE_KEY_PATH", *keyPath))
	}
	if *key != "" {
		fmt.Printf("Direct Key: [provided] (from %s)\n", getValueSource("GITHUB_PRIVATE_KEY", *key))
	}
	fmt.Println()

	// Validate inputs
	if *appIDStr == "" || *installIDStr == "" || (*keyPath == "" && *key == "") {
		fmt.Println("Error: Required parameters missing. Please provide app-id, install-id, and either key or key-path.")
		flag.Usage()
		os.Exit(1)
	}

	// Create spec with the new flattened structure
	spec := &client.Spec{
		Org:            *owner,
		AppID:          *appIDStr,
		InstallationID: *installIDStr,
	}

	// Set private key based on priority
	if *key != "" {
		privateKey, err := sanitizePrivateKey(*key)
		if err != nil {
			fmt.Printf("Error processing private key: %v\n", err)
			os.Exit(1)
		}
		spec.PrivateKey = string(privateKey)
		fmt.Printf("Using private key provided directly (from %s)\n", getValueSource("GITHUB_PRIVATE_KEY", *key))
	} else {
		spec.PrivateKeyPath = *keyPath
		fmt.Printf("Using private key from file: %s (from %s)\n", *keyPath, getValueSource("GITHUB_PRIVATE_KEY_PATH", *keyPath))
	}

	fmt.Println("Creating client with new authentication structure...")

	// Create a context and logger
	ctx := context.Background()
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	// Create client using the new structure
	c, err := client.New(ctx, logger, spec)
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
		fmt.Println("\nTROUBLESHOOTING TIPS:")
		fmt.Println("1. Ensure your private key is a valid PEM-encoded RSA private key")
		fmt.Println("2. The key should start with '-----BEGIN RSA PRIVATE KEY-----' and end with '-----END RSA PRIVATE KEY-----'")
		fmt.Println("3. If your key is base64 encoded, the tool will attempt to decode it")
		fmt.Println("4. Check that your App ID and Installation ID are correct")
		fmt.Println("5. Verify the GitHub App has the necessary permissions for the organization")
		os.Exit(1)
	}

	fmt.Printf("✅ Client created successfully!\n")
	fmt.Printf("App ID: %d\n", c.AppID)
	fmt.Printf("Installation ID: %d\n", c.InstallationID)
	fmt.Printf("Organization: %s\n", c.Org())

	// Create GitHub client directly for testing
	privateKeyBytes := []byte(c.PrivateKey)
	gitHubClient, err := githubinternal.NewGitHubAppClient(ctx, c.AppID, c.InstallationID, privateKeyBytes)
	if err != nil {
		fmt.Printf("❌ Error creating GitHub client: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ GitHub client created successfully!\n\n")

	// Test 1: Check rate limits
	fmt.Println("🔍 Testing API access and rate limits...")
	rateLimits, _, err := gitHubClient.GitHubClient.RateLimit.Get(ctx)
	if err != nil {
		fmt.Printf("❌ Error checking rate limits: %v\n", err)
	} else {
		fmt.Printf("✅ Rate limits retrieved:\n")
		fmt.Printf("  Core: %d/%d (resets at %v)\n",
			rateLimits.Core.Remaining, rateLimits.Core.Limit, rateLimits.Core.Reset.Time)
		fmt.Printf("  Search: %d/%d (resets at %v)\n",
			rateLimits.Search.Remaining, rateLimits.Search.Limit, rateLimits.Search.Reset.Time)
	}

	// Test 2: Check installation permissions (requires JWT authentication, not installation token)
	fmt.Println("\n🔍 Testing installation permissions...")

	// Create a separate JWT-authenticated client for app-level operations
	jwtClient, err := createJWTAuthenticatedClient(c.AppID, privateKeyBytes)
	if err != nil {
		fmt.Printf("❌ Error creating JWT client: %v\n", err)
	} else {
		installation, _, err := jwtClient.Apps.GetInstallation(ctx, c.InstallationID)
		if err != nil {
			fmt.Printf("❌ Error getting installation details: %v\n", err)
			fmt.Println("  This might indicate:")
			fmt.Println("  - Incorrect App ID or Installation ID")
			fmt.Println("  - GitHub App doesn't have access to this installation")
			fmt.Println("  - JWT token generation issue")
		} else {
			fmt.Printf("✅ Installation details:\n")
			fmt.Printf("  Account: %s\n", *installation.Account.Login)
			fmt.Printf("  Target Type: %s\n", *installation.TargetType)
			fmt.Printf("  Created: %v\n", installation.CreatedAt.Time)
			fmt.Printf("  Updated: %v\n", installation.UpdatedAt.Time)

			if installation.Permissions != nil {
				fmt.Printf("  Permissions:\n")
				if installation.Permissions.Metadata != nil {
					fmt.Printf("    Metadata: %s\n", *installation.Permissions.Metadata)
				}
				if installation.Permissions.Contents != nil {
					fmt.Printf("    Contents: %s\n", *installation.Permissions.Contents)
				}
			}
		}
	}

	// Test 3: List accessible repositories
	fmt.Println("\n🔍 Testing repository access...")
	repos, _, err := gitHubClient.GitHubClient.Apps.ListRepos(ctx, nil)
	if err != nil {
		fmt.Printf("❌ Error listing accessible repositories: %v\n", err)
	} else {
		fmt.Printf("✅ App can access %d repositories\n", len(repos.Repositories))
		if len(repos.Repositories) > 0 {
			fmt.Printf("  First few repositories:\n")
			for i, repo := range repos.Repositories {
				if i >= 3 { // Only show first 3
					fmt.Printf("  ... and %d more\n", len(repos.Repositories)-3)
					break
				}
				fmt.Printf("    - %s (private: %v)\n", *repo.FullName, *repo.Private)
			}
		}
	}

	// Test 4: Test organization access
	fmt.Printf("\n🔍 Testing organization access for '%s'...\n", *owner)
	org, _, err := gitHubClient.GitHubClient.Organizations.Get(ctx, *owner)
	if err != nil {
		fmt.Printf("❌ Error accessing organization: %v\n", err)
	} else {
		fmt.Printf("✅ Organization access successful:\n")
		fmt.Printf("  Name: %s\n", *org.Name)
		fmt.Printf("  Public Repos: %d\n", *org.PublicRepos)
		fmt.Printf("  Private Repos: %d\n", *org.TotalPrivateRepos)
	}

	// Test 5: Test repository languages (original test)
	fmt.Printf("\n🔍 Testing repository languages for %s/%s...\n", *owner, *repo)
	langs, err := gitHubClient.GetLanguages(ctx, *owner, *repo)
	if err != nil {
		fmt.Printf("❌ Error fetching languages: %v\n", err)
		fmt.Println("\nPossible issues:")
		fmt.Println("- Repository doesn't exist or isn't accessible")
		fmt.Println("- GitHub App doesn't have 'Contents' permission")
		fmt.Println("- Repository is private and app lacks access")
		os.Exit(1)
	}

	fmt.Printf("✅ Language detection successful!\n")
	fmt.Printf("Repository: %s\n", langs.FullName)
	fmt.Printf("Languages (%d total):\n", len(langs.Languages))
	for _, lang := range langs.Languages {
		fmt.Printf("  - %s\n", lang)
	}

	// Test 6: Performance test - fetch languages for multiple repos
	fmt.Println("\n🔍 Performance test - fetching languages for organization repos...")
	startTime := time.Now()

	orgRepos, _, err := gitHubClient.GitHubClient.Repositories.ListByOrg(ctx, *owner, nil)
	if err != nil {
		fmt.Printf("❌ Error listing org repositories: %v\n", err)
	} else {
		fmt.Printf("✅ Found %d repositories in organization\n", len(orgRepos))

		// Test first 3 repos for performance
		testCount := 3
		if len(orgRepos) < testCount {
			testCount = len(orgRepos)
		}

		successCount := 0
		for i := 0; i < testCount; i++ {
			repo := orgRepos[i]
			if repo.Owner != nil && repo.Owner.Login != nil && repo.Name != nil {
				_, err := gitHubClient.GetLanguages(ctx, *repo.Owner.Login, *repo.Name)
				if err == nil {
					successCount++
				}
			}
		}

		elapsed := time.Since(startTime)
		fmt.Printf("✅ Performance test completed: %d/%d repos processed in %v\n",
			successCount, testCount, elapsed)
	}

	fmt.Println("\n🎉 All authentication tests completed successfully!")
	fmt.Println("Your GitHub App authentication is properly configured.")
}

// getValueSource determines if a value came from CLI flag or environment variable
func getValueSource(envVar, currentValue string) string {
	envValue := os.Getenv(envVar)
	switch envValue {
	case "":
		return "CLI flag (no env var set)"
	case currentValue:
		return "environment variable"
	default:
		return "CLI flag (overriding env var)"
	}
}
