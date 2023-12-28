package services

import (
	"context"
	"fmt"
	"os"

	"github.com/cloudquery/plugin-sdk/v4/transformers"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	gh "github.com/google/go-github/v57/github"
	"github.com/guardian/cq-source-github-languages/internal/github"
)

type Languages struct {
	// TODO find a way to share this with github.go
	FullName  string
	Languages []string
}

func LanguagesTable() *schema.Table {
	return &schema.Table{
		Name:      "github_languages",
		Resolver:  fetchLanguages,
		Transform: transformers.TransformWithStruct(&Languages{}),
	}
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func fetchRepositories(ghClient *gh.Client) ([]*gh.Repository, error) {
	opts := &gh.RepositoryListByOrgOptions{
		ListOptions: gh.ListOptions{
			PerPage: 100,
		}}

	var allRepos []*gh.Repository
	for {
		repos, resp, err := ghClient.Repositories.ListByOrg(context.Background(), "guardian", opts)
		if err != nil {
			return nil, err
		}

		for _, repo := range repos {
			//we are filtering here to only include repos we care about for OKR purposes.
			//the filters can be removed after we are sure we won't hit the rate limit
			if !*repo.Archived && contains(repo.Topics, "production") {
				allRepos = append(allRepos, repo)
			}
		}

		fmt.Println("Counted ", len(allRepos), " repos so far")
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return allRepos, nil
}

func fetchLanguages(ctx context.Context, meta schema.ClientMeta, parent *schema.Resource, res chan<- any) error {
	// TODO authenticate via GitHub App
	token := os.Getenv("GITHUB_ACCESS_TOKEN")
	c := github.CustomClient(token)

	repos, err := fetchRepositories(c.GitHubClient)
	if err != nil {
		return err
	}

	for _, repo := range repos {
		langs, err := c.GetLanguages(*repo.Owner.Login, *repo.Name)
		if err != nil {
			return err
		}

		res <- langs
	}
	return nil
}
