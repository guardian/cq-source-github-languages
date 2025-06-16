package services

import (
	"context"
	"fmt"

	"github.com/cloudquery/plugin-sdk/v4/transformers"

	"github.com/cloudquery/plugin-sdk/v4/schema"
	gh "github.com/google/go-github/v57/github"
	"github.com/guardian/cq-source-github-languages/client"
	"github.com/guardian/cq-source-github-languages/internal/github"
)

func LanguagesTable() *schema.Table {
	return &schema.Table{
		Name:      "github_languages",
		Resolver:  fetchLanguages,
		Transform: transformers.TransformWithStruct(&github.Languages{}),
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
	c := meta.(*client.Client)

	// Initialize GitHub client with App authentication only
	privateKeyBytes := []byte(c.Spec.PrivateKey)
	gitHubClient, err := github.NewGitHubAppClient(c.Spec.AppID, c.Spec.InstallationID, privateKeyBytes)
	if err != nil {
		return fmt.Errorf("failed to create GitHub App client: %w", err)
	}

	// Use the official GitHub client for fetchRepositories
	repos, err := fetchRepositories(gitHubClient.GitHubClient)
	if err != nil {
		return err
	}

	// Use our internal client wrapper for GetLanguages calls
	for _, repo := range repos {
		langs, err := gitHubClient.GetLanguages(*repo.Owner.Login, *repo.Name)
		if err != nil {
			return err
		}

		res <- langs
	}
	return nil
}
