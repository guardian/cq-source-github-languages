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

func TestFilterForValidRepos(t *testing.T) {
	tests := []struct {
		name     string
		repos    []*github.Repository
		expected int
	}{
		{
			name: "filters archived repos",
			repos: []*github.Repository{
				{Archived: &[]bool{true}[0], Topics: []string{"production"}},
				{Archived: &[]bool{false}[0], Topics: []string{"production"}},
			},
			expected: 1,
		},
		{
			name: "filters repos without production topic",
			repos: []*github.Repository{
				{Archived: &[]bool{false}[0], Topics: []string{"experimental"}},
				{Archived: &[]bool{false}[0], Topics: []string{"production"}},
			},
			expected: 1,
		},
		{
			name: "handles nil archived field",
			repos: []*github.Repository{
				{Archived: nil, Topics: []string{"production"}}, // This might cause a panic
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterForValidRepos(tt.repos)
			if len(result) != tt.expected {
				t.Errorf("filterForValidRepos() = %d repos, want %d", len(result), tt.expected)
			}
		})
	}
}
