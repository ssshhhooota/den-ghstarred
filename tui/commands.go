package tui

import (
	"context"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	gh "github.com/ssshhhooota/den-ghstarred/github"
)

func (m Model) fetchRepos() tea.Cmd {
	return func() tea.Msg {
		repos, err := m.client.ListStarredRepos(context.Background(), false)
		return ReposLoadedMsg{Repos: repos, Err: err}
	}
}

func (m Model) forceReloadRepos() tea.Cmd {
	return func() tea.Msg {
		repos, err := m.client.ListStarredRepos(context.Background(), true)
		return ReposLoadedMsg{Repos: repos, Err: err}
	}
}

func (m Model) fetchReadme(owner, repo string) tea.Cmd {
	fullName := owner + "/" + repo
	return func() tea.Msg {
		content, err := m.client.GetReadme(context.Background(), owner, repo)
		return ReadmeLoadedMsg{Content: content, Err: err, FullName: fullName}
	}
}

func (m Model) searchRepos(query string) tea.Cmd {
	return func() tea.Msg {
		repos, err := m.client.SearchRepos(context.Background(), query)
		if err != nil {
			return SearchResultMsg{Repos: repos, Err: err}
		}
		// シンプルな単語クエリの場合、org 名としても検索してマージ
		if !strings.Contains(query, " ") && !strings.Contains(query, ":") {
			orgRepos, orgErr := m.client.SearchRepos(context.Background(), "org:"+query)
			if orgErr == nil {
				seen := make(map[string]bool, len(repos))
				for _, r := range repos {
					seen[r.FullName] = true
				}
				for _, r := range orgRepos {
					if !seen[r.FullName] {
						repos = append(repos, r)
					}
				}
			}
		}
		return SearchResultMsg{Repos: repos}
	}
}

func (m Model) toggleStar(repo *gh.Repo, star bool) tea.Cmd {
	return func() tea.Msg {
		var err error
		if star {
			err = m.client.StarRepo(context.Background(), repo.Owner, repo.Name)
		} else {
			err = m.client.UnstarRepo(context.Background(), repo.Owner, repo.Name)
		}
		return StarToggleMsg{Repo: repo, Starred: star, Err: err}
	}
}
