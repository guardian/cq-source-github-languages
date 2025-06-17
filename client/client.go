package client

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/rs/zerolog"
)

type Client struct {
	logger         zerolog.Logger
	Spec           Spec
	appID          int64
	installationID int64
	privateKey     string
}

func (c *Client) ID() string {
	return "github-languages"
}

func (c *Client) Logger() *zerolog.Logger {
	return &c.logger
}

func (c *Client) AppID() int64 {
	return c.appID
}

func (c *Client) InstallationID() int64 {
	return c.installationID
}

func (c *Client) PrivateKey() string {
	return c.privateKey
}

func (c *Client) Org() string {
	if c.Spec.AppAuth != nil {
		return c.Spec.AppAuth.Org
	}
	return ""
}

func New(ctx context.Context, logger zerolog.Logger, s *Spec) (Client, error) {
	var appID, installationID int64
	var privateKeyContent string

	// Validate that app_auth is provided
	if s.AppAuth == nil {
		return Client{}, fmt.Errorf("app_auth configuration is required")
	}

	// Convert string app_id to int64
	if s.AppAuth.AppID != "" {
		var err error
		appID, err = strconv.ParseInt(s.AppAuth.AppID, 10, 64)
		if err != nil {
			return Client{}, fmt.Errorf("failed to parse app_id '%s' as integer: %w", s.AppAuth.AppID, err)
		}
	}

	// Convert string installation_id to int64
	if s.AppAuth.InstallationID != "" {
		var err error
		installationID, err = strconv.ParseInt(s.AppAuth.InstallationID, 10, 64)
		if err != nil {
			return Client{}, fmt.Errorf("failed to parse installation_id '%s' as integer: %w", s.AppAuth.InstallationID, err)
		}
	}

	// Handle private key - prefer content over file path
	if s.AppAuth.PrivateKey != "" {
		privateKeyContent = s.AppAuth.PrivateKey
	} else if s.AppAuth.PrivateKeyPath != "" {
		keyBytes, err := os.ReadFile(s.AppAuth.PrivateKeyPath)
		if err != nil {
			return Client{}, fmt.Errorf("failed to read private key from file %s: %w", s.AppAuth.PrivateKeyPath, err)
		}
		privateKeyContent = string(keyBytes)
	}

	// Validate required parameters
	if appID == 0 {
		return Client{}, fmt.Errorf("github app id is required")
	}
	if installationID == 0 {
		return Client{}, fmt.Errorf("github app installation id is required")
	}
	if privateKeyContent == "" {
		return Client{}, fmt.Errorf("github app private key is required")
	}

	c := Client{
		logger:         logger,
		Spec:           *s,
		appID:          appID,
		installationID: installationID,
		privateKey:     privateKeyContent,
	}

	return c, nil
}
