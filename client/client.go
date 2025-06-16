package client

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"
)

type Client struct {
	logger zerolog.Logger
	Spec   Spec
}

func (c *Client) ID() string {
	return "github-languages"
}

func (c *Client) Logger() *zerolog.Logger {
	return &c.logger
}

func New(ctx context.Context, logger zerolog.Logger, s *Spec) (Client, error) {
	// Validate GitHub App authentication parameters
	if s.AppID == 0 {
		return Client{}, fmt.Errorf("github app id is required")
	}
	if s.InstallationID == 0 {
		return Client{}, fmt.Errorf("github app installation id is required")
	}
	if s.PrivateKey == "" {
		return Client{}, fmt.Errorf("github app private key is required")
	}

	c := Client{
		logger: logger,
		Spec:   *s,
	}

	return c, nil
}
