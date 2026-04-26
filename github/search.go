package github

import (
	"context"
	"fmt"

	"github.com/google/go-github/v60/github"
)

// SearchRepos searches GitHub globally; returns at most 30 results.
func (c *Client) SearchRepos(ctx context.Context, query string) ([]*Repo, error) {
	result, _, err := c.gh.Search.Repositories(ctx, query, &github.SearchOptions{
		ListOptions: github.ListOptions{PerPage: SearchLimit},
	})
	if err != nil {
		return nil, fmt.Errorf("search repos: %w", err)
	}
	if result == nil {
		return nil, fmt.Errorf("search repos: empty response")
	}
	repos := make([]*Repo, 0, len(result.Repositories))
	for _, r := range result.Repositories {
		repos = append(repos, &Repo{
			Owner:       r.GetOwner().GetLogin(),
			Name:        r.GetName(),
			FullName:    r.GetFullName(),
			Description: r.GetDescription(),
			StarCount:   r.GetStargazersCount(),
			Language:    r.GetLanguage(),
		})
	}
	return repos, nil
}

// StarRepo stars owner/repo for the authenticated user.
func (c *Client) StarRepo(ctx context.Context, owner, repo string) error {
	_, err := c.gh.Activity.Star(ctx, owner, repo)
	if err != nil {
		return fmt.Errorf("star repo: %w", err)
	}
	return nil
}

// UnstarRepo removes the star from owner/repo.
func (c *Client) UnstarRepo(ctx context.Context, owner, repo string) error {
	_, err := c.gh.Activity.Unstar(ctx, owner, repo)
	if err != nil {
		return fmt.Errorf("unstar repo: %w", err)
	}
	return nil
}
