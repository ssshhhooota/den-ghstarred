package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	gh "github.com/ssshhhooota/den-ghstarred/github"
	"github.com/ssshhhooota/den-ghstarred/tui"
)

func main() {
	token, err := gh.GetToken()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	client := gh.NewClient(token)
	m := tui.New(client)

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
