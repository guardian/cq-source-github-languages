package client

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/rs/zerolog"
)

type Client struct {
	logger         zerolog.Logger
	Spec           Spec
	AppID          int64
	InstallationID int64
	PrivateKey     string
}

func (c *Client) ID() string {
	return "github-languages"
}

func (c *Client) Logger() *zerolog.Logger {
	return &c.logger
}

func (c *Client) Org() string {
	return c.Spec.Org
}

func New(ctx context.Context, logger zerolog.Logger, s *Spec) (Client, error) {
	var appID, installationID int64
	var privateKeyContent string

	// Validate required fields
	if s.Org == "" {
		return Client{}, fmt.Errorf("organization is required")
	}

	if s.AppID != "" {
		// Handle potential file interpolation syntax
		appIDStr := strings.TrimSpace(s.AppID)
		if strings.HasPrefix(appIDStr, "${file:") && strings.HasSuffix(appIDStr, "}") {
			logger.Warn().Msg("app_id appears to contain file interpolation syntax - ensure CloudQuery has processed this correctly")
		}

		var err error
		appID, err = strconv.ParseInt(appIDStr, 10, 64)
		if err != nil {
			return Client{}, fmt.Errorf("failed to parse app_id '%s' as integer: %w", appIDStr, err)
		}
	}

	if s.InstallationID != "" {
		// Handle potential file interpolation syntax
		installIDStr := strings.TrimSpace(s.InstallationID)
		if strings.HasPrefix(installIDStr, "${file:") && strings.HasSuffix(installIDStr, "}") {
			logger.Warn().Msg("installation_id appears to contain file interpolation syntax - ensure CloudQuery has processed this correctly")
		}

		var err error
		installationID, err = strconv.ParseInt(installIDStr, 10, 64)
		if err != nil {
			return Client{}, fmt.Errorf("failed to parse installation_id '%s' as integer: %w", installIDStr, err)
		}
	}

	if s.PrivateKeyPath != "" {
		keyBytes, err := os.ReadFile(s.PrivateKeyPath)
		if err != nil {
			return Client{}, fmt.Errorf("failed to read private key from file %s: %w", s.PrivateKeyPath, err)
		}
		privateKeyContent = strings.TrimSpace(string(keyBytes))
		logger.Info().Str("key_path", s.PrivateKeyPath).Msg("loaded private key from file")
	} else if s.PrivateKey != "" {
		privateKeyContent = strings.TrimSpace(s.PrivateKey)
		logger.Info().Msg("using private key from config")
	}

	if appID == 0 {
		return Client{}, fmt.Errorf("github app id is required")
	}
	if installationID == 0 {
		return Client{}, fmt.Errorf("github app installation id is required")
	}
	if privateKeyContent == "" {
		return Client{}, fmt.Errorf("github app private key is required (either private_key or private_key_path)")
	}

	if !strings.Contains(privateKeyContent, "-----BEGIN") || !strings.Contains(privateKeyContent, "-----END") {
		return Client{}, fmt.Errorf("private key must be in PEM format with proper BEGIN/END markers")
	}

	logger.Info().
		Int64("app_id", appID).
		Int64("installation_id", installationID).
		Str("org", s.Org).
		Msg("GitHub App client configured successfully")

	return Client{
		logger:         logger,
		Spec:           *s,
		AppID:          appID,
		InstallationID: installationID,
		PrivateKey:     privateKeyContent,
	}, nil
}
