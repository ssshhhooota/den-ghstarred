package tui

import (
	"testing"

	gh "github.com/ssshhhooota/den-ghstarred/github"
)

func TestSelectedRepo_Empty(t *testing.T) {
	m := Model{}
	if got := m.selectedRepo(); got != nil {
		t.Errorf("selectedRepo() = %v, want nil", got)
	}
}

func TestSelectedRepo_ValidCursor(t *testing.T) {
	repo := &gh.Repo{FullName: "foo/bar"}
	m := Model{
		starred: StarredPane{
			repos:  []*gh.Repo{repo},
			cursor: 0,
		},
	}
	got := m.selectedRepo()
	if got == nil || got.FullName != "foo/bar" {
		t.Errorf("selectedRepo() = %v, want foo/bar", got)
	}
}

func TestCurrentRepo_StarredPane(t *testing.T) {
	repo := &gh.Repo{FullName: "foo/bar"}
	m := Model{
		sidebarFocused:  true,
		activePanelLeft: 0,
		starred: StarredPane{
			repos:  []*gh.Repo{repo},
			cursor: 0,
		},
	}
	got := m.currentRepo()
	if got == nil || got.FullName != "foo/bar" {
		t.Errorf("currentRepo() = %v, want foo/bar", got)
	}
}

func TestCurrentRepo_SearchPane(t *testing.T) {
	repo := &gh.Repo{FullName: "baz/qux"}
	m := Model{
		sidebarFocused:  true,
		activePanelLeft: 1,
		search: SearchPane{
			results: []*gh.Repo{repo},
			cursor:  0,
			focused: false,
		},
	}
	got := m.currentRepo()
	if got == nil || got.FullName != "baz/qux" {
		t.Errorf("currentRepo() = %v, want baz/qux", got)
	}
}

func TestCurrentRepo_SearchPaneInputFocused(t *testing.T) {
	repo := &gh.Repo{FullName: "baz/qux"}
	m := Model{
		sidebarFocused:  true,
		activePanelLeft: 1,
		search: SearchPane{
			results: []*gh.Repo{repo},
			cursor:  0,
			focused: true, // 入力フィールドにフォーカス → nil
		},
	}
	got := m.currentRepo()
	if got != nil {
		t.Errorf("currentRepo() = %v, want nil when search input focused", got)
	}
}

func TestSelectedRepo_OutOfBounds(t *testing.T) {
	repos := []*gh.Repo{{FullName: "foo/bar"}}
	m := Model{
		starred: StarredPane{
			repos:  repos,
			cursor: len(repos), // 範囲外
		},
	}
	if got := m.selectedRepo(); got != nil {
		t.Errorf("selectedRepo() = %v, want nil for out-of-bounds cursor", got)
	}
}

func TestCurrentRepo_ReadmePanel(t *testing.T) {
	repo := &gh.Repo{FullName: "foo/bar"}
	m := Model{
		sidebarFocused: false, // README パネルにフォーカス
		starred:        StarredPane{repos: []*gh.Repo{repo}},
		readme:         ReadmePanel{fullName: "foo/bar"},
	}
	got := m.currentRepo()
	if got == nil || got.FullName != "foo/bar" {
		t.Errorf("currentRepo() = %v, want foo/bar when readme panel focused", got)
	}
}

func TestRepoByFullName(t *testing.T) {
	starred := &gh.Repo{FullName: "foo/bar"}
	searched := &gh.Repo{FullName: "baz/qux"}
	m := Model{
		starred: StarredPane{repos: []*gh.Repo{starred}},
		search:  SearchPane{results: []*gh.Repo{searched}},
	}

	if got := m.repoByFullName("foo/bar"); got != starred {
		t.Errorf("repoByFullName(foo/bar) = %v, want %v", got, starred)
	}
	if got := m.repoByFullName("baz/qux"); got != searched {
		t.Errorf("repoByFullName(baz/qux) = %v, want %v", got, searched)
	}
	if got := m.repoByFullName("unknown/repo"); got != nil {
		t.Errorf("repoByFullName(unknown/repo) = %v, want nil", got)
	}
}
