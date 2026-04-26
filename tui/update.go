package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	gh "github.com/ssshhhooota/den-ghstarred/github"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if !m.loading {
			newRightW := m.width - leftPanelWidth() - 4
			if newRightW < 1 {
				newRightW = 1
			}
			m.readme.vp.Width = newRightW
			m.readme.vp.Height = m.rightPanelContentH()
			if m.readme.raw != "" {
				lines := renderMarkdown(m.readme.raw, newRightW)
				m.readme.lines = lines
				m.readme.vp.SetContent(strings.Join(lines, "\n"))
			}
		}
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case ReposLoadedMsg:
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.starred.repos = msg.Repos
		m.starred.visualItems = buildVisualItems(m.starred.repos)
		m.starredIndex = make(map[string]bool, len(m.starred.repos))
		for _, r := range m.starred.repos {
			m.starredIndex[r.FullName] = true
		}
		rightW := m.width - leftPanelWidth() - 4
		if rightW < 1 {
			rightW = 1
		}
		m.readme.vp = viewport.New(rightW, m.rightPanelContentH())
		m, autoCmd := m.withAutoLoadReadme()
		if autoCmd == nil {
			m.readme.vp.SetContent("← Select a repository and press Enter to load README")
		}
		return m, autoCmd

	case ReadmeLoadedMsg:
		m.readme.cursor = 0
		m.readme.fullName = msg.FullName
		if msg.Err != nil {
			m.readme.raw = ""
			m.readme.lines = nil
			m.readme.vp.SetContent(fmt.Sprintf("Error: %v", msg.Err))
			m.readme.vp.GotoTop()
			return m, nil
		}
		if msg.Content == "" {
			m.readme.raw = ""
			m.readme.lines = nil
			m.readme.vp.SetContent("(no README found)")
			m.readme.vp.GotoTop()
			return m, nil
		}
		m.readme.raw = msg.Content
		lines := renderMarkdown(msg.Content, m.readme.vp.Width)
		m.readme.lines = lines
		m.readme.vp.SetContent(strings.Join(lines, "\n"))
		m.readme.vp.GotoTop()
		return m, nil

	case SearchResultMsg:
		m.search.loading = false
		if msg.Err != nil {
			m.search.err = msg.Err.Error()
			m.search.results = nil
			m.search.focused = true
			m.search.input.Focus()
		} else {
			m.search.results = msg.Repos
			m.search.cursor = 0
			m.search.offset = 0
			m.search.err = ""
			if len(msg.Repos) == 0 {
				m.search.focused = true
				m.search.input.Focus()
			} else {
				m.search.focused = false
				m.search.input.Blur()
				return m.withAutoLoadReadme()
			}
		}
		return m, nil

	case StarToggleMsg:
		if msg.Err != nil {
			m.infoMsg = "Error: " + msg.Err.Error()
		} else if msg.Starred {
			m.infoMsg = "★ Starred " + msg.Repo.FullName
			m.starred.repos = append([]*gh.Repo{msg.Repo}, m.starred.repos...)
			m.starredIndex[msg.Repo.FullName] = true
			m.starred.visualItems = buildVisualItems(m.starred.repos)
			gh.SaveCache(m.starred.repos)
			// Compensate for prepend: cursor shifts by 1 to stay on same repo
			m.starred.cursor++
		} else {
			m.infoMsg = "☆ Unstarred " + msg.Repo.FullName
			for i, r := range m.starred.repos {
				if r.FullName == msg.Repo.FullName {
					if i < m.starred.cursor {
						m.starred.cursor--
					}
					break
				}
			}
			newRepos := make([]*gh.Repo, 0, len(m.starred.repos)-1)
			for _, r := range m.starred.repos {
				if r.FullName != msg.Repo.FullName {
					newRepos = append(newRepos, r)
				}
			}
			m.starred.repos = newRepos
			delete(m.starredIndex, msg.Repo.FullName)
			m.starred.visualItems = buildVisualItems(m.starred.repos)
			if cv := m.cursorVisualIdx(); m.starred.offset > cv {
				m.starred.offset = cv
			}
			gh.SaveCache(m.starred.repos)
			if m.starred.cursor >= len(m.starred.repos) && m.starred.cursor > 0 {
				m.starred.cursor = len(m.starred.repos) - 1
			}
		}
		return m, tea.Tick(infoMsgDuration, func(time.Time) tea.Msg { return clearInfoMsg{} })

	case clearInfoMsg:
		m.infoMsg = ""
		return m, nil

	case debouncedSearchCursorMsg:
		if m.sidebarFocused && m.activePanelLeft == 1 && msg.cursor == m.search.cursor && !m.search.focused {
			if msg.cursor < len(m.search.results) {
				r := m.search.results[msg.cursor]
				m.readme.raw = ""
				m.readme.lines = nil
				m.readme.vp.SetContent("Loading preview...")
				return m, m.fetchReadme(r.Owner, r.Name)
			}
		}
		return m, nil

	case debouncedCursorMsg:
		if msg.cursor == m.starred.cursor {
			if repo := m.selectedRepo(); repo != nil {
				m.readme.vp.SetContent("Loading README...")
				return m, m.fetchReadme(repo.Owner, repo.Name)
			}
		}
		return m, nil

	case tea.KeyMsg:
		if m.sidebarFocused && m.activePanelLeft == 1 {
			return m.handleSearchPaneKey(msg)
		}
		return m.handleNormalKey(msg)
	}

	var cmd tea.Cmd
	m.readme.vp, cmd = m.readme.vp.Update(msg)
	return m, cmd
}
