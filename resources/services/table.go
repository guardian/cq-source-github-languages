package services

import (
	"context"
	"fmt"

	"github.com/apache/arrow/go/v14/arrow"
	"github.com/cloudquery/plugin-sdk/v4/schema"
	"github.com/guardian/cq-source-github-languages/client"
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
	cl := meta.(*client.Client)
	return fmt.Errorf("not implemented. client id: " + cl.ID())
}
