package tui

import (
	"time"

	"github.com/charmbracelet/lipgloss"
)

const (
	debounceDelay   = 300 * time.Millisecond
	infoMsgDuration = 3 * time.Second
)

var (
	// colorPrimary is the accent color used for focus indicators and highlights.
	colorPrimary = lipgloss.Color("#93C5FD")
	// colorBorder is used for panel borders in unfocused state.
	colorBorder = lipgloss.Color("#F4F4F5")
	// colorMuted is used for secondary text (owner labels, separators).
	colorMuted = lipgloss.Color("#D4D4D8")
	// colorForeground is used for selected item text and unfocused panel headers.
	// Matches colorBorder intentionally — both are light neutral tones.
	colorForeground = lipgloss.Color("#F4F4F5")
)
