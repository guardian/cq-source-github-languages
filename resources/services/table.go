package services

import (
	"context"

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
				Name: "id",
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
	customClient := github.NewClient("")
	l, err := c.GetLanguages("guardian", "service-catalogue")
	if err != nil {
		return err
	}

}
