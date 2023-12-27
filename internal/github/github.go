package github

import (
	"context"
	"fmt"

	"github.com/google/go-github/v57/github"
)

type Languages struct {
	FullName  string
	Languages string
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
		Languages: fmt.Sprint(langs),
	}
	return l, nil

}
