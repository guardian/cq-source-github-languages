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
	token string
}

func NewClient(baseURL string) *Client {
	return &Client{
		token: "",
	}
}

func (c *Client) GetLanguages(owner string, name string) (*Languages, error) {
	client := github.NewClient(nil).WithAuthToken(c.token)
	langs, _, err := client.Repositories.ListLanguages(context.Background(), owner, name)
	if err != nil {
		return nil, err
	}
	l := &Languages{
		FullName:  owner + "/" + name,
		Languages: fmt.Sprint(langs),
	}
	return l, nil

}
