# GitHub App Authentication Test Tool

A comprehensive testing utility for validating GitHub App authentication configuration before running the CloudQuery sync process.

## Overview

The `test-auth` tool performs a series of authentication tests to ensure your GitHub App is properly configured and has the necessary permissions to access your organisation's repositories. This helps identify and resolve authentication issues early, preventing failures during the actual CloudQuery sync.

## Features

- **Multiple Authentication Tests**: Validates both JWT and installation token authentication
- **Rate Limit Monitoring**: Shows remaining API calls and reset times
- **Permission Verification**: Checks GitHub App installation permissions
- **Repository Access Testing**: Lists accessible repositories and tests language detection
- **Performance Testing**: Benchmarks language fetching across multiple repositories
- **Comprehensive Error Reporting**: Provides detailed error messages with troubleshooting guidance
- **Flexible Configuration**: Supports both environment variables and command-line flags

## Building the Tool

```bash
make build-test-auth
```

This creates a `test-auth` binary in the project root.

## Configuration

The tool supports multiple ways to provide authentication credentials, with command-line flags taking precedence over environment variables.

### Environment Variables

```bash
export GITHUB_APP_ID="your_app_id"
export GITHUB_INSTALLATION_ID="your_installation_id"
export GITHUB_PRIVATE_KEY_PATH="/path/to/private-key.pem"
# OR
export GITHUB_PRIVATE_KEY="-----BEGIN RSA PRIVATE KEY-----..."
```

### Command-Line Flags

```bash
./test-auth \
  -app-id "123456" \
  -install-id "12345678" \
  -key-path "/path/to/private-key.pem" \
  -owner "your-org" \
  -repo "test-repo"
```

## Usage Examples

### Basic Authentication Test

```bash
# Using environment variables
./test-auth

# Using command-line flags
./test-auth -app-id "123456" -install-id "12345678" -key-path "/path/to/key.pem"
```

### Test Specific organisation/Repository

```bash
./test-auth -owner "guardian" -repo "frontend"
```

### Using Direct Private Key

```bash
./test-auth \
  -app-id "123456" \
  -install-id "12345678" \
  -key "-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA...
-----END RSA PRIVATE KEY-----"
```

## Command-Line Options

| Flag | Environment Variable | Description |
|------|---------------------|-------------|
| `-app-id` | `GITHUB_APP_ID` | GitHub App ID |
| `-install-id` | `GITHUB_INSTALLATION_ID` | GitHub App Installation ID |
| `-key-path` | `GITHUB_PRIVATE_KEY_PATH` | Path to private key file |
| `-key` | `GITHUB_PRIVATE_KEY` | Private key content directly |
| `-owner` | N/A | Repository owner/organisation (default: "guardian") |
| `-repo` | N/A | Repository name for testing (default: "cq-source-github-languages") |

## Test Sequence

The tool performs the following tests in order:

### 1. 🔍 Client Configuration
- Validates input parameters
- Creates and configures the authentication client
- Parses and validates private key format

### 2. 🔍 API Access and Rate Limits
- Tests basic API connectivity
- Shows current rate limit status
- Displays reset times for different API endpoints

### 3. 🔍 Installation Permissions
- Retrieves GitHub App installation details
- Shows installation permissions (metadata, contents, etc.)
- Validates app has access to the specified installation

### 4. 🔍 Repository Access
- Lists repositories accessible to the GitHub App
- Shows first few repositories with their visibility status
- Counts total accessible repositories

### 5. 🔍 organisation Access
- Tests organisation-level API access
- Retrieves organisation details and repository counts
- Validates organisation permissions

### 6. 🔍 Language Detection
- Tests repository language detection on a specific repository
- Validates the core functionality used by the CloudQuery plugin
- Shows detected languages for the test repository

### 7. 🔍 Performance Testing
- Fetches languages for multiple organization repositories
- Measures processing time and success rate
- Provides performance metrics

## Output Example

```
Configuration:
App ID: 123456 (from environment variable)
Installation ID: 12345678 (from CLI flag)
Key Path: /path/to/key.pem (from environment variable)

✅ Client created successfully!
App ID: 123456
Installation ID: 12345678
Organization: guardian

✅ GitHub client created successfully!

🔍 Testing API access and rate limits...
✅ Rate limits retrieved:
  Core: 4850/5000 (resets at 2024-01-15 14:30:00)
  Search: 30/30 (resets at 2024-01-15 13:35:00)

🔍 Testing installation permissions...
✅ Installation details:
  Account: guardian
  Target Type: Organization
  Created: 2024-01-10 10:15:30
  Updated: 2024-01-15 12:20:45
  Permissions:
    Metadata: read
    Contents: read

🎉 All authentication tests completed successfully!
Your GitHub App authentication is properly configured.
```

## Private Key Formats

The tool supports multiple private key formats:

### RSA PKCS#1 Format
```
-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA...
-----END RSA PRIVATE KEY-----
```

### PKCS#8 Format
```
-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC...
-----END PRIVATE KEY-----
```

### Base64 Encoded Keys
The tool automatically detects and decodes base64-encoded private keys.

## Troubleshooting

### Common Issues

#### JWT Token Decode Error
```
❌ Error getting installation details: GET https://api.github.com/app/installations/12345: 401 A JSON web token could not be decoded
```

**Solutions:**
- Verify App ID is correct
- Check private key format (must be PEM-encoded RSA key)
- Ensure private key matches the GitHub App
- Verify system clock is synchronized (JWT tokens are time-sensitive)

#### Installation Not Found
```
❌ Error creating GitHub client: failed to create installation token (HTTP 404)
```

**Solutions:**
- Verify Installation ID is correct
- Check that the GitHub App is installed on the target organization
- Ensure the App ID matches the installed GitHub App

#### Permission Denied
```
❌ Error listing accessible repositories: GET https://api.github.com/installation/repositories: 403 Forbidden
```

**Solutions:**
- Verify GitHub App has necessary permissions (Contents: Read, Metadata: Read)
- Check that the app is installed with repository access
- Ensure the organization allows the GitHub App

#### Private Key Format Issues
```
❌ Error processing private key: unsupported private key format
```

**Solutions:**
- Ensure key is in PEM format with proper headers/footers
- Convert key to RSA PKCS#1 or PKCS#8 format if necessary
- Check for extra whitespace or encoding issues

### Debug Information

The tool provides extensive debug output including:
- Private key format detection
- JWT token claims and timing
- API response details
- Rate limit information
- Installation permissions

## Integration with CloudQuery

Once `test-auth` completes successfully, you can use the same credentials with CloudQuery:

```yaml
kind: source
spec:
  name: "github-languages"
  path: "guardian/github-languages"
  version: "v1.0.0"
  destinations:
    - "postgresql"
  spec:
    org: "guardian"
    app_id: "${GITHUB_APP_ID}"
    installation_id: "${GITHUB_INSTALLATION_ID}"
    private_key_path: "${GITHUB_PRIVATE_KEY_PATH}"
```

## Exit Codes

- `0`: All tests passed successfully
- `1`: Configuration error or test failure

## Requirements

- Go 1.24.3 or later
- Valid GitHub App with appropriate permissions
- Network access to GitHub API (api.github.com)

## Support

If you encounter issues not covered in this README:

1. Run `test-auth` with verbose output to get detailed error messages
2. Check GitHub App configuration and permissions
3. Verify network connectivity to GitHub API
4. Ensure system time is synchronized for JWT token validation
