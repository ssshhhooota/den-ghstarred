package tui

import gh "github.com/ssshhhooota/den-ghstarred/github"

// ReposLoadedMsg is sent when the starred repos list has been fetched.
type ReposLoadedMsg struct {
	Repos []*gh.Repo
	Err   error
}

// ReadmeLoadedMsg is sent when a repository's README has been fetched.
type ReadmeLoadedMsg struct {
	Content  string
	Err      error
	FullName string
}

// SearchResultMsg is sent when a GitHub search completes.
type SearchResultMsg struct {
	Repos []*gh.Repo
	Err   error
}

// StarToggleMsg is sent when a star/unstar API call completes.
type StarToggleMsg struct {
	Repo    *gh.Repo
	Starred bool
	Err     error
}

// clearInfoMsg clears the status message after infoMsgDuration.
type clearInfoMsg struct{}

// debouncedCursorMsg triggers a README load for the starred pane after debounce.
type debouncedCursorMsg struct{ cursor int }

// debouncedSearchCursorMsg triggers a README preview load for the search pane after debounce.
type debouncedSearchCursorMsg struct{ cursor int }
