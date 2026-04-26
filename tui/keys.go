package tui

import (
	"os/exec"
	"runtime"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	gh "github.com/ssshhhooota/den-ghstarred/github"
)

func (m Model) handleSearchPaneKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "*":
		if !m.search.focused && m.search.cursor >= 0 && m.search.cursor < len(m.search.results) {
			repo := m.search.results[m.search.cursor]
			return m, m.toggleStar(repo, !m.starredIndex[repo.FullName])
		}
		return m, nil

	case "tab":
		if len(m.search.results) > 0 {
			m.search.focused = !m.search.focused
			if m.search.focused {
				m.search.input.Focus()
			} else {
				m.search.input.Blur()
			}
		}
		return m, nil

	case "?":
		if !m.search.focused {
			m.showHelp = !m.showHelp
			m.readme.vp.Height = m.rightPanelContentH()
			return m, nil
		}

	case "q", "ctrl+c":
		if !m.search.focused {
			return m, tea.Quit
		}
		var cmd tea.Cmd
		m.search.input, cmd = m.search.input.Update(msg)
		return m, cmd

	case "esc":
		m.activePanelLeft = 0
		m.search.input.Blur()
		return m.withAutoLoadReadme()

	case "1":
		if !m.search.focused {
			m.activePanelLeft = 0
			m.search.input.Blur()
			return m.withAutoLoadReadme()
		}
		var cmd tea.Cmd
		m.search.input, cmd = m.search.input.Update(msg)
		return m, cmd

	case "enter":
		if m.search.focused {
			query := strings.TrimSpace(m.search.input.Value())
			if query == "" {
				return m, nil
			}
			m.search.loading = true
			m.search.results = nil
			m.search.err = ""
			return m, m.searchRepos(query)
		}
		// Enter on a result: focus right panel (skip fetch if preview already loaded)
		if len(m.search.results) > 0 && m.search.cursor >= 0 && m.search.cursor < len(m.search.results) {
			selected := m.search.results[m.search.cursor]
			m.sidebarFocused = false
			if m.readme.fullName != selected.FullName {
				m.readme.lines = nil
				m.readme.cursor = 0
				m.readme.vp.SetContent("Loading README...")
				return m, m.fetchReadme(selected.Owner, selected.Name)
			}
		}
		return m, nil

	case "o":
		if m.search.focused {
			var cmd tea.Cmd
			m.search.input, cmd = m.search.input.Update(msg)
			return m, cmd
		}
		if m.search.cursor < len(m.search.results) {
			openInBrowser("https://github.com/" + m.search.results[m.search.cursor].FullName)
		}
		return m, nil

	case "j", "down":
		if m.search.focused {
			var cmd tea.Cmd
			m.search.input, cmd = m.search.input.Update(msg)
			return m, cmd
		}
		if m.search.cursor < len(m.search.results)-1 {
			m.search.cursor++
			listH := m.searchListH()
			if m.search.cursor >= m.search.offset+listH {
				m.search.offset = m.search.cursor - listH + 1
			}
			cur := m.search.cursor
			return m, tea.Tick(debounceDelay, func(time.Time) tea.Msg {
				return debouncedSearchCursorMsg{cursor: cur}
			})
		}
		return m, nil

	case "k", "up":
		if m.search.focused {
			var cmd tea.Cmd
			m.search.input, cmd = m.search.input.Update(msg)
			return m, cmd
		}
		if m.search.cursor > 0 {
			m.search.cursor--
			if m.search.cursor < m.search.offset {
				m.search.offset = m.search.cursor
			}
			cur := m.search.cursor
			return m, tea.Tick(debounceDelay, func(time.Time) tea.Msg {
				return debouncedSearchCursorMsg{cursor: cur}
			})
		}
		return m, nil

	default:
		if m.search.focused {
			var cmd tea.Cmd
			m.search.input, cmd = m.search.input.Update(msg)
			return m, cmd
		}
		return m, nil
	}
	return m, nil
}

func (m Model) handleNormalKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "?":
		m.showHelp = !m.showHelp
		m.readme.vp.Height = m.rightPanelContentH()
		return m, nil

	case "q", "ctrl+c":
		return m, tea.Quit

	case "1":
		m.sidebarFocused = true
		m.activePanelLeft = 0
		return m.withAutoLoadReadme()

	case "2":
		m.sidebarFocused = true
		m.activePanelLeft = 1
		if len(m.search.results) == 0 {
			m.search.focused = true
			m.search.input.Focus()
		}
		return m.withAutoLoadReadme()

	case "R":
		m.loading = true
		m.readme.raw = ""
		m.readme.lines = nil
		m.readme.cursor = 0
		m.readme.fullName = ""
		m.starred.repos = nil
		m.starred.cursor = 0
		m.starred.offset = 0
		m.sidebarFocused = true
		m.activePanelLeft = 0
		return m, tea.Batch(m.spinner.Tick, m.forceReloadRepos())

	case "*":
		if repo := m.currentRepo(); repo != nil {
			return m, m.toggleStar(repo, !m.starredIndex[repo.FullName])
		}
		return m, nil

	case "enter":
		if m.sidebarFocused && m.activePanelLeft == 0 {
			m.sidebarFocused = false
			repo := m.selectedRepo()
			if repo != nil && m.readme.fullName != repo.FullName {
				m.readme.lines = nil
				m.readme.cursor = 0
				m.readme.vp.SetContent("Loading README...")
				return m, m.fetchReadme(repo.Owner, repo.Name)
			}
		}

	case "esc":
		m.sidebarFocused = true

	case "o":
		if repo := m.currentRepo(); repo != nil {
			openInBrowser("https://github.com/" + repo.FullName)
		}

	case "j", "down":
		if m.sidebarFocused && m.activePanelLeft == 0 {
			if m.starred.cursor < len(m.starred.repos)-1 {
				m.starred.cursor++
				listH := m.starredListH()
				cv := m.cursorVisualIdx()
				if cv >= m.starred.offset+listH {
					newOffset := cv - listH + 1
					if newOffset < 0 {
						newOffset = 0
					}
					m.starred.offset = newOffset
				}
				cur := m.starred.cursor
				return m, tea.Tick(debounceDelay, func(time.Time) tea.Msg {
					return debouncedCursorMsg{cursor: cur}
				})
			}
		} else if !m.sidebarFocused && len(m.readme.lines) > 0 && m.readme.cursor < len(m.readme.lines)-1 {
			m.readme.cursor++
			if m.readme.cursor >= m.readme.vp.YOffset+m.readme.vp.Height {
				m.readme.vp.LineDown(1)
			}
		}

	case "k", "up":
		if m.sidebarFocused && m.activePanelLeft == 0 {
			if m.starred.cursor > 0 {
				m.starred.cursor--
				cv := m.cursorVisualIdx()
				if cv < m.starred.offset {
					m.starred.offset = cv
				}
				cur := m.starred.cursor
				return m, tea.Tick(debounceDelay, func(time.Time) tea.Msg {
					return debouncedCursorMsg{cursor: cur}
				})
			}
		} else if !m.sidebarFocused && m.readme.cursor > 0 {
			m.readme.cursor--
			if m.readme.cursor < m.readme.vp.YOffset {
				m.readme.vp.LineUp(1)
			}
		}

	case "d", " ":
		if !m.sidebarFocused {
			m.readme.vp.HalfViewDown()
			m.readme.cursor = m.readme.vp.YOffset
		}

	case "u":
		if !m.sidebarFocused {
			m.readme.vp.HalfViewUp()
			m.readme.cursor = m.readme.vp.YOffset
		}
	}
	return m, nil
}

// withAutoLoadReadme loads the README for the currently focused repo if not already loaded.
func (m Model) withAutoLoadReadme() (Model, tea.Cmd) {
	var repo *gh.Repo
	if m.activePanelLeft == 0 {
		repo = m.selectedRepo()
	} else {
		if !m.search.focused && m.search.cursor < len(m.search.results) {
			repo = m.search.results[m.search.cursor]
		}
	}
	if repo == nil || repo.FullName == m.readme.fullName {
		return m, nil
	}
	m.readme.lines = nil
	m.readme.cursor = 0
	m.readme.vp.SetContent("Loading README...")
	return m, m.fetchReadme(repo.Owner, repo.Name)
}

func openInBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	cmd.Start() //nolint:errcheck
}
