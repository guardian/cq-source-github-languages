package services

import (
	"context"
	"os"

	"github.com/cloudquery/plugin-sdk/v4/transformers"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	"github.com/guardian/cq-source-github-languages/internal/github"
)

type Languages struct {
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

func fetchLanguages(ctx context.Context, meta schema.ClientMeta, parent *schema.Resource, res chan<- any) error {
	token := os.Getenv("GITHUB_TOKEN")
	c := github.CustomClient(token)
	allRepos, _, err := c.GitHubClient.Repositories.ListByOrg(ctx, "guardian", nil)
	if err != nil {
		return err
	}
	for _, repo := range allRepos[1:10] {
		langs, err := c.GetLanguages(*repo.Owner.Login, *repo.Name)
		if err != nil {
			return err
		}

		res <- langs

	}
	return nil
}
