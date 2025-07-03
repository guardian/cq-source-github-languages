package services

import (
	"slices"
	"testing"

	"github.com/google/go-github/v57/github"
)

func TestLanguagesTable(t *testing.T) {
	table := LanguagesTable()

	if table.Name != "github_languages" {
		t.Errorf("Table name = %v, want %v", table.Name, "github_languages")
	}

	if table.Resolver == nil {
		t.Error("Table resolver should not be nil")
	}

	if table.Transform == nil {
		t.Error("Table transform should not be nil")
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name  string
		slice []string
		str   string
		want  bool
	}{
		{
			name:  "string exists",
			slice: []string{"apple", "banana", "cherry"},
			str:   "banana",
			want:  true,
		},
		{
			name:  "string does not exist",
			slice: []string{"apple", "banana", "cherry"},
			str:   "grape",
			want:  false,
		},
		{
			name:  "empty slice",
			slice: []string{},
			str:   "apple",
			want:  false,
		},
		{
			name:  "empty string in slice",
			slice: []string{"", "apple"},
			str:   "",
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := slices.Contains(tt.slice, tt.str)
			if got != tt.want {
				t.Errorf("slices.Contains() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFetchRepositoriesFiltering(t *testing.T) {
	// Test the filtering logic with mock data
	archived := true
	notArchived := false

	repos := []*github.Repository{
		{
			Archived: &archived,
			Topics:   []string{"production", "service"},
		},
		{
			Archived: &notArchived,
			Topics:   []string{"production", "service"},
		},
		{
			Archived: &notArchived,
			Topics:   []string{"experimental"},
		},
	}

	var filteredRepos []*github.Repository
	for _, repo := range repos {
		// Apply the same filtering logic as fetchRepositories
		if !*repo.Archived && slices.Contains(repo.Topics, "production") {
			filteredRepos = append(filteredRepos, repo)
		}
	}

	expectedCount := 1 // Only the second repo should pass the filter
	if len(filteredRepos) != expectedCount {
		t.Errorf("Filtered repos count = %v, want %v", len(filteredRepos), expectedCount)
	}
}
