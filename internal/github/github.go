package github

import (
	"context"

	"github.com/google/go-github/v57/github"
	"golang.org/x/exp/maps"
)

type Languages struct {
	// TODO find a way to share this with table.go
	FullName  string
	Languages []string
}

type Client struct {
	GitHubClient *github.Client
}

func CustomClient(token string) *Client {
	return &Client{
		GitHubClient: github.NewClient(nil).WithAuthToken(token),
	}
}

func (c *Client) GetLanguages(owner string, name string) (*Languages, error) {
	langs, _, err := c.GitHubClient.Repositories.ListLanguages(context.Background(), owner, name)
	if err != nil {
		return nil, err
	}
	l := &Languages{
		FullName:  owner + "/" + name,
		Languages: maps.Keys(langs),
	}
	return l, nil

}
