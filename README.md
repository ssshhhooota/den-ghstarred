# den-ghstarred

A TUI tool to browse your GitHub starred repositories in the terminal.

## Features

- List starred repositories grouped by owner
- Markdown preview of README files
- Open repositories in a browser
- 1-hour cache for API responses

## Installation

```sh
go install github.com/ssshhhooota/den-ghstarred@latest
```

Or build locally:

```sh
git clone https://github.com/ssshhhooota/den-ghstarred
cd den-ghstarred
make install
```

## Authentication

If you are already logged in with [GitHub CLI](https://cli.github.com/), authentication happens automatically.

```sh
gh auth login
```

Or set the environment variable manually:

```sh
export GITHUB_TOKEN=ghp_...
```

## Usage

```sh
den-ghstarred
```

## Development

```sh
make run    # run
make build  # build
make fmt    # format
make lint   # lint
```

## Dependencies

- [charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea) — TUI framework
- [charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss) — styling
- [charmbracelet/glamour](https://github.com/charmbracelet/glamour) — markdown rendering
- [google/go-github](https://github.com/google/go-github) — GitHub API client
