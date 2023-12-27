package github

import (
	"io"
	"net/http"
)

type Languages struct {
	FullName  string
	Languages string
}

const defaultURL = "https://api.github.com"

type Client struct {
	baseURL string
	client  *http.Client
	token   string
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: defaultURL,
		client:  http.DefaultClient,
		token:   "",
	}
}

func (c *Client) GetLanguages(owner string, name string) (*Languages, error) {
	languagesURL := c.baseURL + "/repos/" + owner + "/" + name + "/languages"
	req, err := http.NewRequest("GET", languagesURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "cloudquery")
	req.Header.Set("Authorization", "token "+c.token)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	l := &Languages{
		FullName:  owner + "/" + name,
		Languages: string(body),
	}
	return l, nil

}
