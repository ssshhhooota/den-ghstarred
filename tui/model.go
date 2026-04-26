package tui

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	gh "github.com/ssshhhooota/den-ghstarred/github"
)

// StarredPane holds state for the [1] Starred Repos pane.
type StarredPane struct {
	repos       []*gh.Repo
	cursor      int
	offset      int
	visualItems []visualItem
}

// SearchPane holds state for the [2] Search pane.
type SearchPane struct {
	results []*gh.Repo
	cursor  int
	offset  int
	input   textinput.Model
	focused bool // true = input focused, false = results list focused
	loading bool
	err     string
}

// ReadmePanel holds state for the right README panel.
type ReadmePanel struct {
	vp       viewport.Model
	raw      string
	lines    []string
	cursor   int
	fullName string // FullName of the currently displayed repo
}

// Model holds the application state.
type Model struct {
	client  *gh.Client
	width   int
	height  int
	loading bool
	err     error
	spinner spinner.Model

	starred StarredPane
	search  SearchPane
	readme  ReadmePanel

	activePanelLeft int // 0 = Starred, 1 = Search
	sidebarFocused  bool
	infoMsg         string
	starredIndex    map[string]bool
	showHelp        bool
}

// New creates a new Model.
func New(client *gh.Client) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot

	ti := textinput.New()
	ti.Placeholder = "Filter starred repos..."
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(colorMuted)
	ti.Prompt = "  "
	ti.CharLimit = 200
	ti.Width = leftPanelWidth() - 4

	return Model{
		client:         client,
		loading:        true,
		spinner:        s,
		search:         SearchPane{input: ti, focused: true},
		sidebarFocused: true,
		starredIndex:   make(map[string]bool),
	}
}

// Init kicks off the initial starred repos fetch.
func (m Model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.fetchRepos())
}

// footerH returns the number of terminal rows occupied by the footer.
func (m Model) footerH() int {
	if m.showHelp {
		return 3
	}
	return 2
}

// mainH returns the usable terminal rows excluding the footer.
func (m Model) mainH() int {
	return m.height - m.footerH()
}

// rightPanelContentH returns the content height for the right README panel.
func (m Model) rightPanelContentH() int {
	return m.mainH() - 2
}

// searchListH returns the number of visible result rows in the [2] pane.
func (m Model) searchListH() int {
	mainH := m.mainH()
	paneH1 := (mainH - 4) / 2
	if paneH1 < 1 {
		paneH1 = 1
	}
	paneH2 := mainH - 4 - paneH1
	if paneH2 < 1 {
		paneH2 = 1
	}
	listH := paneH2 - 3
	if listH < 1 {
		listH = 1
	}
	return listH
}

// starredListH returns the number of visible list rows in the [1] pane.
func (m Model) starredListH() int {
	mainH := m.mainH()
	idealPaneH1 := (mainH - 4) / 2
	if idealPaneH1 < 1 {
		idealPaneH1 = 1
	}
	paneH1 := idealPaneH1
	if contentH := len(m.starred.visualItems) + 2; contentH < paneH1 {
		paneH1 = contentH
	}
	listH := paneH1 - 2
	if listH < 1 {
		listH = 1
	}
	return listH
}

// cursorVisualIdx returns the visual row index for m.starred.cursor.
func (m Model) cursorVisualIdx() int {
	for i, item := range m.starred.visualItems {
		if !item.isOwner && item.repoIdx == m.starred.cursor {
			return i
		}
	}
	return 0
}

// selectedRepo returns the repo at the current starred pane cursor.
func (m Model) selectedRepo() *gh.Repo {
	if len(m.starred.repos) == 0 || m.starred.cursor < 0 || m.starred.cursor >= len(m.starred.repos) {
		return nil
	}
	return m.starred.repos[m.starred.cursor]
}

// repoByFullName searches starred and search results for a matching repo.
func (m Model) repoByFullName(fullName string) *gh.Repo {
	for _, r := range m.starred.repos {
		if r.FullName == fullName {
			return r
		}
	}
	for _, r := range m.search.results {
		if r.FullName == fullName {
			return r
		}
	}
	return nil
}

// currentRepo returns the repo currently in focus (sidebar or readme panel).
func (m Model) currentRepo() *gh.Repo {
	switch {
	case m.sidebarFocused && m.activePanelLeft == 0:
		return m.selectedRepo()
	case m.sidebarFocused && m.activePanelLeft == 1:
		if !m.search.focused && m.search.cursor < len(m.search.results) {
			return m.search.results[m.search.cursor]
		}
	case !m.sidebarFocused:
		return m.repoByFullName(m.readme.fullName)
	}
	return nil
}
