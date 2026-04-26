package tui

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
	gh "github.com/ssshhhooota/den-ghstarred/github"
)

var (
	stylePanelBorder = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorBorder)

	stylePanelBorderFocused = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorPrimary)

	styleSelected = lipgloss.NewStyle().
			Foreground(colorForeground).
			Bold(true)

	styleMuted = lipgloss.NewStyle().Foreground(colorMuted)

	styleMainCursor = lipgloss.NewStyle().
			Background(lipgloss.Color("#1E3A8A")).
			Foreground(lipgloss.Color("#F8FAFC"))

	styleSearchSelected = lipgloss.NewStyle().Foreground(colorForeground).Bold(true)
	styleSearchStar     = lipgloss.NewStyle().Foreground(lipgloss.Color("#FACC15"))
)

func leftPanelWidth() int { return 35 }

// View renders the full TUI.
func (m Model) View() string {
	if m.width == 0 {
		return ""
	}
	if m.loading {
		return fmt.Sprintf("\n  %s Loading starred repos...", m.spinner.View())
	}
	if m.err != nil {
		return fmt.Sprintf("\n  Error: %v\n\n  Press q to quit.", m.err)
	}

	footerH := 2 // border-top (1) + content line (1)
	if m.showHelp {
		footerH = 3 // border-top (1) + 2 content lines
	}
	mainH := m.height - footerH
	footer := renderFooter(m, m.width)

	leftW := leftPanelWidth()
	rightW := m.width - leftW - 4

	// Two left panes stack vertically. Each pane: content + 2 border rows.
	// paneH1 + paneH2 = mainH - 4; distribute remainder to bottom pane.
	idealPaneH1 := (mainH - 4) / 2
	if idealPaneH1 < 1 {
		idealPaneH1 = 1
	}
	// コンテンツが少ない場合は starred ペインを縮小して空白行を出さない
	paneH1 := idealPaneH1
	if contentH := len(m.starred.visualItems) + 2; contentH < paneH1 {
		paneH1 = contentH
	}
	paneH2 := mainH - 4 - paneH1
	if paneH2 < 1 {
		paneH2 = 1
		paneH1 = mainH - 4 - 1
	}

	starredBorder := stylePanelBorder
	searchBorder := stylePanelBorder
	if m.sidebarFocused && m.activePanelLeft == 0 {
		starredBorder = stylePanelBorderFocused
	} else if m.sidebarFocused && m.activePanelLeft == 1 {
		searchBorder = stylePanelBorderFocused
	}

	styledStarred := starredBorder.Width(leftW).Height(paneH1).Render(renderStarredPane(m, leftW, paneH1))
	styledSearch := searchBorder.Width(leftW).Height(paneH2).Render(renderSearchPane(m, leftW, paneH2))
	styledLeft := lipgloss.JoinVertical(lipgloss.Left, styledStarred, styledSearch)

	rightBorder := stylePanelBorder
	borderFg := colorBorder
	if !m.sidebarFocused {
		rightBorder = stylePanelBorderFocused
		borderFg = colorPrimary
	}
	styledRight := rightBorder.Width(rightW).Height(mainH - 2).Render(renderRight(m, rightW, mainH-2))
	if m.readme.fullName != "" {
		styledRight = addBorderTitle(styledRight, m.readme.fullName, borderFg, rightW)
	}

	main := lipgloss.JoinHorizontal(lipgloss.Top, styledLeft, styledRight)
	return lipgloss.JoinVertical(lipgloss.Left, main, footer)
}

var styleFooterKey = lipgloss.NewStyle().Foreground(colorPrimary).Bold(true)
var styleFooterDesc = lipgloss.NewStyle().Foreground(colorMuted)
var styleFooterSep = lipgloss.NewStyle().Foreground(colorBorder)

func renderFooter(m Model, width int) string {
	sep := styleFooterSep.Render(" / ")
	hint := func(key, desc string) string {
		return styleFooterKey.Render(key) + styleFooterDesc.Render(" "+desc)
	}
	style := lipgloss.NewStyle().
		Width(width).
		BorderTop(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(colorBorder)

	switch {
	case m.sidebarFocused && m.activePanelLeft == 1 && m.search.focused:
		// Search input — ?キーは文字入力になるため常にcompact
		return style.Render(strings.Join([]string{
			hint("type", "filter"),
			hint("enter", "search"),
			hint("tab", "results"),
			hint("esc", "back"),
		}, sep))

	case m.sidebarFocused && m.activePanelLeft == 1:
		// Search results
		compact := strings.Join([]string{
			hint("j/k", "navigate"),
			hint("enter", "README"),
			hint("o", "browser"),
			hint("*", "star"),
			hint("tab", "input"),
			hint("1", "starred"),
			hint("esc", "back"),
			hint("?", "help"),
		}, sep)
		if m.showHelp {
			row2 := strings.Join([]string{
				hint("enter", "open README"),
				hint("o", "open in browser"),
				hint("*", "toggle star"),
				hint("tab", "back to input"),
				hint("1", "switch to starred"),
				hint("?", "close help"),
			}, sep)
			return style.Render(compact + "\n" + row2)
		}
		return style.Render(compact)

	case !m.sidebarFocused:
		// README panel
		compact := strings.Join([]string{
			hint("j/k", "scroll"),
			hint("d/spc", "page↓"),
			hint("u", "page↑"),
			hint("o", "browser"),
			hint("*", "star"),
			hint("esc", "sidebar"),
			hint("q", "quit"),
			hint("?", "help"),
		}, sep)
		if m.showHelp {
			row2 := strings.Join([]string{
				hint("j/k ↑/↓", "line scroll"),
				hint("d/space", "half page down"),
				hint("u", "half page up"),
				hint("o", "open in browser"),
				hint("*", "star/unstar"),
				hint("esc", "back to sidebar"),
				hint("?", "close help"),
			}, sep)
			return style.Render(compact + "\n" + row2)
		}
		return style.Render(compact)

	default:
		// Starred pane
		compact := strings.Join([]string{
			hint("j/k", "navigate"),
			hint("enter", "README"),
			hint("o", "browser"),
			hint("*", "star"),
			hint("2", "search"),
			hint("R", "reload"),
			hint("q", "quit"),
			hint("?", "help"),
		}, sep)
		if m.showHelp {
			row2 := strings.Join([]string{
				hint("enter", "open README"),
				hint("o", "open in browser"),
				hint("*", "toggle star"),
				hint("2", "switch to search"),
				hint("R", "reload all"),
				hint("q", "quit"),
				hint("?", "close help"),
			}, sep)
			return style.Render(compact + "\n" + row2)
		}
		return style.Render(compact)
	}
}

func renderPaneHeader(title string, focused bool) string {
	if focused {
		return lipgloss.NewStyle().Bold(true).Foreground(colorPrimary).Render(title)
	}
	return lipgloss.NewStyle().Bold(true).Foreground(colorForeground).Render(title)
}

func renderStarredPane(m Model, width, height int) string {
	focused := m.sidebarFocused && m.activePanelLeft == 0
	sep := styleMuted.Render(strings.Repeat("-", width))
	listH := height - 2
	if listH < 1 {
		listH = 1
	}

	lines := []string{renderPaneHeader("[1] Starred Repos", focused), sep}
	items := m.starred.visualItems

	start := m.starred.offset
	if start >= len(items) {
		start = 0
	}
	// 最下部でアイテムが listH に満たないとき start を前に詰めて空白行をなくす
	if maxStart := len(items) - listH; maxStart > 0 && start > maxStart {
		start = maxStart
	}

	for i := start; i < len(items) && i < start+listH; i++ {
		item := items[i]
		if item.isOwner {
			lines = append(lines, styleMuted.Render(truncate(item.owner+"/", width)))
		} else {
			repo := m.starred.repos[item.repoIdx]
			label := truncate(repo.Name, width-4)
			if item.repoIdx == m.starred.cursor {
				if focused {
					lines = append(lines, styleSelected.Render("  > "+label))
				} else {
					lines = append(lines, "  > "+label)
				}
			} else {
				lines = append(lines, "    "+label)
			}
		}
	}
	if len(lines) > height {
		lines = lines[:height]
	}
	return strings.Join(lines, "\n")
}

func renderRight(m Model, width, height int) string {
	var lines []string
	if m.infoMsg != "" {
		lines = append(lines, lipgloss.NewStyle().Foreground(colorPrimary).Bold(true).Width(width).Render(m.infoMsg))
		height--
	}
	if len(m.readme.lines) == 0 {
		vpLines := strings.Split(m.readme.vp.View(), "\n")
		if len(vpLines) > height {
			vpLines = vpLines[:height]
		}
		lines = append(lines, strings.Join(vpLines, "\n"))
		return strings.Join(lines, "\n")
	}
	cursorStyle := styleMainCursor.Width(width)
	start := m.readme.vp.YOffset
	end := start + height
	if end > len(m.readme.lines) {
		end = len(m.readme.lines)
	}
	for i := start; i < end; i++ {
		line := m.readme.lines[i]
		if i == m.readme.cursor {
			lines = append(lines, cursorStyle.Render(stripANSI(line)))
		} else {
			lines = append(lines, line)
		}
	}
	return strings.Join(lines, "\n")
}

type visualItem struct {
	isOwner bool
	owner   string
	repoIdx int
}

func buildVisualItems(repos []*gh.Repo) []visualItem {
	type ownerGroup struct {
		name    string
		indices []int
	}
	var groups []ownerGroup
	ownerPos := map[string]int{}
	for i, repo := range repos {
		if pos, ok := ownerPos[repo.Owner]; ok {
			groups[pos].indices = append(groups[pos].indices, i)
		} else {
			ownerPos[repo.Owner] = len(groups)
			groups = append(groups, ownerGroup{name: repo.Owner, indices: []int{i}})
		}
	}
	var items []visualItem
	for _, g := range groups {
		items = append(items, visualItem{isOwner: true, owner: g.name})
		for _, idx := range g.indices {
			items = append(items, visualItem{repoIdx: idx})
		}
	}
	return items
}

// addBorderTitle injects a repository name into the top border line of a
// rendered lipgloss panel. It replaces the dash segment after "╭" with
// " title " followed by the remaining dashes, keeping the total visual width
// unchanged. borderFg is the color to apply to the border characters.
func addBorderTitle(rendered, title string, borderFg lipgloss.Color, panelWidth int) string {
	if title == "" {
		return rendered
	}
	nl := strings.Index(rendered, "\n")
	if nl < 0 {
		return rendered
	}
	topLine := rendered[:nl]
	rest := rendered[nl:]

	totalWidth := lipgloss.Width(topLine)
	// Truncate title so it fits: ╭ [space title space] [at least 2 dashes] ╮
	maxTitle := totalWidth - 6 // "╭ " + " " + "--" + "╮"
	if maxTitle < 1 {
		return rendered
	}
	title = truncate(title, maxTitle)

	titlePart := " " + title + " "
	needed := 2 + lipgloss.Width(titlePart) // "╭" + titlePart + "╮"
	dashCount := totalWidth - needed
	if dashCount < 0 {
		return rendered
	}

	borderSty := lipgloss.NewStyle().Foreground(borderFg)
	newTop := borderSty.Render("╭") +
		lipgloss.NewStyle().Bold(true).Foreground(colorForeground).Render(titlePart) +
		borderSty.Render(strings.Repeat("─", dashCount)) +
		borderSty.Render("╮")

	return newTop + rest
}

var ansiEscape = regexp.MustCompile(`\x1b\[[0-9;]*[mGKHF]`)

func stripANSI(s string) string { return ansiEscape.ReplaceAllString(s, "") }

func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	if maxLen <= 1 {
		return "…"
	}
	return string(runes[:maxLen-1]) + "…"
}

// renderSearchPane renders the [2] Search pane content.
func renderSearchPane(m Model, width, height int) string {
	focused := m.sidebarFocused && m.activePanelLeft == 1
	header := renderPaneHeader("[2] Search", focused)
	inputLine := m.search.input.View()
	sep := styleMuted.Render(strings.Repeat("-", width))

	// header + inputLine + sep = 3 fixed lines
	listH := height - 3
	if listH < 1 {
		listH = 1
	}

	lines := []string{header, inputLine, sep}

	switch {
	case m.search.loading:
		lines = append(lines, "  "+m.spinner.View()+" Searching...")
	case m.search.err != "":
		lines = append(lines, styleMuted.Render("  Error: "+m.search.err))
	case m.search.focused && m.search.input.Value() != "":
		lines = append(lines, styleMuted.Render("  Press Enter to filter"))
	case len(m.search.results) == 0 && m.search.input.Value() == "":
		lines = append(lines, styleMuted.Render("  Type to filter starred repos"))
	case len(m.search.results) == 0:
		lines = append(lines, styleMuted.Render("  (no results)"))
	default:
		start := m.search.offset
		end := start + listH
		if end > len(m.search.results) {
			end = len(m.search.results)
		}
		for i := start; i < end; i++ {
			r := m.search.results[i]
			star := "  "
			if m.starredIndex[r.FullName] {
				star = styleSearchStar.Render("★ ")
			}
			label := truncate(r.FullName, width-6)
			line := star + label
			if i == m.search.cursor && !m.search.focused {
				if focused {
					lines = append(lines, styleSearchSelected.Render("  > "+line))
				} else {
					lines = append(lines, "  > "+line)
				}
			} else {
				lines = append(lines, "    "+line)
			}
		}
	}

	if len(lines) > height {
		lines = lines[:height]
	}
	return strings.Join(lines, "\n")
}
