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

func fetchRepositories(ctx context.Context, ghClient *gh.Client, org string) ([]*gh.Repository, error) {
	opts := &gh.RepositoryListByOrgOptions{
		ListOptions: gh.ListOptions{
			PerPage: 100,
		}}

	var allRepos []*gh.Repository
	for {
		repos, resp, err := ghClient.Repositories.ListByOrg(ctx, org, opts)
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
	c, ok := meta.(*client.Client)
	if !ok {
		return fmt.Errorf("failed to assert meta as *client.Client")
	}

	logger := c.Logger()
	logger.Info().Msg("starting language fetch process")

	// Initialize GitHub client with App authentication
	privateKeyBytes := []byte(c.PrivateKey)
	gitHubClient, err := github.NewGitHubAppClient(ctx, c.AppID, c.InstallationID, privateKeyBytes)
	if err != nil {
		logger.Error().Err(err).Msg("failed to create GitHub App client")
		return fmt.Errorf("failed to create GitHub App client: %w", err)
	}

	logger.Info().Str("org", c.Org()).Msg("fetching repositories")

	// Use the official GitHub client for fetchRepositories
	repos, err := fetchRepositories(ctx, gitHubClient.GitHubClient, c.Org())
	if err != nil {
		logger.Error().Err(err).Str("org", c.Org()).Msg("failed to fetch repositories")
		return fmt.Errorf("failed to fetch repositories for org %s: %w", c.Org(), err)
	}

	logger.Info().Int("repo_count", len(repos)).Msg("fetched repositories, now getting languages")

	// Use our internal client wrapper for GetLanguages calls
	for i, repo := range repos {
		if repo.Owner == nil || repo.Owner.Login == nil || repo.Name == nil {
			logger.Warn().Int("repo_index", i).Msg("skipping repository with missing owner or name")
			continue
		}

		langs, err := gitHubClient.GetLanguages(ctx, *repo.Owner.Login, *repo.Name)
		if err != nil {
			logger.Error().
				Err(err).
				Str("owner", *repo.Owner.Login).
				Str("repo", *repo.Name).
				Msg("failed to get languages for repository")
			return fmt.Errorf("failed to get languages for %s/%s: %w", *repo.Owner.Login, *repo.Name, err)
		}

		logger.Debug().
			Str("repo", langs.FullName).
			Int("language_count", len(langs.Languages)).
			Msg("fetched languages for repository")

		res <- langs
	}

	logger.Info().Int("total_repos", len(repos)).Msg("completed language fetch process")
	return nil
}
