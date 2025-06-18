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

	if s.AppID != "" {
		var err error
		appID, err = strconv.ParseInt(s.AppID, 10, 64)
		if err != nil {
			return Client{}, fmt.Errorf("failed to parse app_id '%s' as integer: %w", s.AppID, err)
		}
	}

	if s.InstallationID != "" {
		var err error
		installationID, err = strconv.ParseInt(s.InstallationID, 10, 64)
		if err != nil {
			return Client{}, fmt.Errorf("failed to parse installation_id '%s' as integer: %w", s.InstallationID, err)
		}
	}

	if s.PrivateKeyPath != "" {
		keyBytes, err := os.ReadFile(s.PrivateKeyPath)
		if err != nil {
			return Client{}, fmt.Errorf("failed to read private key from file %s: %w", s.PrivateKeyPath, err)
		}
		privateKeyContent = string(keyBytes)
	} else if s.PrivateKey != "" {
		privateKeyContent = s.PrivateKey
	}

	if appID == 0 {
		return Client{}, fmt.Errorf("github app id is required")
	}
	if installationID == 0 {
		return Client{}, fmt.Errorf("github app installation id is required")
	}
	if privateKeyContent == "" {
		return Client{}, fmt.Errorf("github app private key is required")
	}

	if !strings.Contains(privateKeyContent, "-----BEGIN") || !strings.Contains(privateKeyContent, "-----END") {
		return Client{}, fmt.Errorf("private key must be in PEM format")
	}

	return Client{
		logger:         logger,
		Spec:           *s,
		AppID:          appID,
		InstallationID: installationID,
		PrivateKey:     privateKeyContent,
	}, nil
}
