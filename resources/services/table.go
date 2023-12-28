package services

import (
	"context"
	"os"

	"github.com/apache/arrow/go/v14/arrow"
	"github.com/cloudquery/plugin-sdk/v4/schema"
	"github.com/guardian/cq-source-github-languages/internal/github"
)

func LanguagesTable() *schema.Table {
	return &schema.Table{
		Name:     "github-languages",
		Resolver: fetchLanguages,
		Columns: []schema.Column{
			{
				Name: "full_name",
				Type: arrow.BinaryTypes.String,
			},
			{
				Name: "languages",
				Type: arrow.BinaryTypes.String,
			},
		},
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
