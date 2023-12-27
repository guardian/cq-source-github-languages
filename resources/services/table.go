package services

import (
	"context"
	"fmt"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	"github.com/cloudquery/plugin-sdk/v4/transformers"
	"github.com/guardian/cq-source-github-languages/client"
	"github.com/guardian/cq-source-github-languages/internal/github"
)

func LanguagesTable() *schema.Table {
	return &schema.Table{
		Name:     "github-languages",
		Resolver: fetchLanguages,
		Columns:  transformers.TransformWithStruct(&github.Languages{}),
	}
}

func fetchLanguages(ctx context.Context, meta schema.ClientMeta, parent *schema.Resource, res chan<- any) error {
	cl := meta.(*client.Client)
	return fmt.Errorf("not implemented. client id: " + cl.ID())
}
