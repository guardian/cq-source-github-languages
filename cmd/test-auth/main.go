package main

import (
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/guardian/cq-source-github-languages/internal/github"
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

	// Ensure key has proper BEGIN and END markers with newlines
	if !strings.Contains(key, "-----BEGIN") {
		fmt.Println("Warning: Private key does not contain BEGIN marker")
	}
	if !strings.Contains(key, "-----END") {
		fmt.Println("Warning: Private key does not contain END marker")
	}

	return []byte(key), nil
}

func main() {
	// Define flags for command line arguments
	appIDStr := flag.String("app-id", os.Getenv("GITHUB_APP_ID"), "GitHub App ID (or set GITHUB_APP_ID env)")
	installIDStr := flag.String("install-id", os.Getenv("GITHUB_INSTALLATION_ID"), "GitHub Installation ID (or set GITHUB_INSTALLATION_ID env)")
	keyPath := flag.String("key-path", os.Getenv("GITHUB_PRIVATE_KEY_PATH"), "Path to GitHub App private key file (or set GITHUB_PRIVATE_KEY_PATH env)")
	key := flag.String("key", os.Getenv("GITHUB_PRIVATE_KEY"), "GitHub App private key contents directly (or set GITHUB_PRIVATE_KEY env)")
	owner := flag.String("owner", "guardian", "GitHub repository owner/organization")
	repo := flag.String("repo", "cq-source-github-languages", "GitHub repository name")

	flag.Parse()

	// Validate inputs
	if *appIDStr == "" || *installIDStr == "" || (*keyPath == "" && *key == "") {
		fmt.Println("Error: Required parameters missing. Please provide app-id, install-id, and either key or key-path.")
		flag.Usage()
		os.Exit(1)
	}

	// Convert string IDs to int64
	appID, err := strconv.ParseInt(*appIDStr, 10, 64)
	if err != nil {
		fmt.Printf("Error parsing app ID: %v\n", err)
		os.Exit(1)
	}

	installID, err := strconv.ParseInt(*installIDStr, 10, 64)
	if err != nil {
		fmt.Printf("Error parsing installation ID: %v\n", err)
		os.Exit(1)
	}

	// Determine which key to use - prefer direct key if provided
	var privateKey []byte
	if *key != "" {
		privateKey, err = sanitizePrivateKey(*key)
		if err != nil {
			fmt.Printf("Error processing private key: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Using private key provided directly via --key flag or GITHUB_PRIVATE_KEY env var")
	} else {
		// Read private key file
		keyData, err := os.ReadFile(*keyPath)
		if err != nil {
			fmt.Printf("Error reading private key file: %v\n", err)
			os.Exit(1)
		}
		privateKey, err = sanitizePrivateKey(string(keyData))
		if err != nil {
			fmt.Printf("Error processing private key from file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Using private key from file: %s\n", *keyPath)
	}

	fmt.Println("Authenticating with GitHub App...")
	fmt.Printf("App ID: %d\n", appID)
	fmt.Printf("Installation ID: %d\n", installID)

	// Create a context for the API calls
	ctx := context.Background()

	// Create GitHub client
	client, err := github.NewGitHubAppClient(ctx, appID, installID, privateKey)
	if err != nil {
		fmt.Printf("Error creating GitHub client: %v\n", err)
		fmt.Println("\nTROUBLESHOOTING TIPS:")
		fmt.Println("1. Ensure your private key is a valid PEM-encoded RSA private key")
		fmt.Println("2. The key should start with '-----BEGIN RSA PRIVATE KEY-----' and end with '-----END RSA PRIVATE KEY-----'")
		fmt.Println("3. If your key is base64 encoded, the tool will attempt to decode it")
		fmt.Println("4. Check that your App ID and Installation ID are correct")
		os.Exit(1)
	}

	fmt.Printf("Fetching languages for %s/%s...\n", *owner, *repo)

	// Test by fetching languages for a repository
	langs, err := client.GetLanguages(ctx, *owner, *repo)
	if err != nil {
		fmt.Printf("Error fetching languages: %v\n", err)
		os.Exit(1)
	}

	// Print results
	fmt.Println("Authentication successful!")
	fmt.Printf("Repository: %s\n", langs.FullName)
	fmt.Println("Languages:")
	for _, lang := range langs.Languages {
		fmt.Printf("- %s\n", lang)
	}
}
