package github

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/google/go-github/v60/github"
	"golang.org/x/oauth2"
)

// Client wraps the GitHub API client.
type Client struct {
	gh *github.Client
}

// GetToken retrieves a GitHub token via `gh auth token`, falling back to GITHUB_TOKEN.
func GetToken() (string, error) {
	out, err := exec.Command("gh", "auth", "token").Output()
	if err == nil {
		if token := strings.TrimSpace(string(out)); token != "" {
			return token, nil
		}
	}
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		return token, nil
	}
	return "", fmt.Errorf("GitHub token not found: run 'gh auth login' or set GITHUB_TOKEN")
}

// NewClient returns a Client authenticated with the given token.
func NewClient(token string) *Client {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(context.Background(), ts)
	return &Client{gh: github.NewClient(tc)}
}

// Repo is a simplified representation of a GitHub repository.
type Repo struct {
	Owner       string
	Name        string
	FullName    string
	Description string
	StarCount   int
	Language    string
}

// ListStarredRepos returns the authenticated user's starred repositories.
// Results are cached to disk for 1 hour to reduce API requests.
// Pass force=true to bypass the cache and re-fetch from the API.
func (c *Client) ListStarredRepos(ctx context.Context, force bool) ([]*Repo, error) {
	if force {
		if err := clearCache(); err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("clear cache: %w", err)
		}
	}
	if repos, ok := loadCache(); ok {
		return repos, nil
	}
	opts := &github.ActivityListStarredOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	var all []*Repo
	for {
		starred, resp, err := c.gh.Activity.ListStarred(ctx, "", opts)
		if err != nil {
			return nil, fmt.Errorf("list starred repos: %w", err)
		}
		for _, s := range starred {
			r := s.GetRepository()
			all = append(all, &Repo{
				Owner:       r.GetOwner().GetLogin(),
				Name:        r.GetName(),
				FullName:    r.GetFullName(),
				Description: r.GetDescription(),
				StarCount:   r.GetStargazersCount(),
				Language:    r.GetLanguage(),
			})
		}
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	saveCache(all)
	return all, nil
}

// GetReadme fetches the README content of owner/repo as plain text.
// Returns empty string if no README exists.
func (c *Client) GetReadme(ctx context.Context, owner, repo string) (string, error) {
	readme, _, err := c.gh.Repositories.GetReadme(ctx, owner, repo, nil)
	if err != nil {
		var ghErr *github.ErrorResponse
		if errors.As(err, &ghErr) && ghErr.Response.StatusCode == http.StatusNotFound {
			return "", nil
		}
		return "", fmt.Errorf("get readme: %w", err)
	}
	content, err := readme.GetContent()
	if err != nil {
		return "", fmt.Errorf("decode readme: %w", err)
	}
	return content, nil
}
