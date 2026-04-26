package tui

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/glamour"
	xansi "github.com/charmbracelet/x/ansi"
)

var (
	reHTMLTag    = regexp.MustCompile(`<[^>]+>`)
	reHTMLEntity = regexp.MustCompile(`&[a-zA-Z]+;|&#[0-9]+;`)
	reImage      = regexp.MustCompile(`!\[[^\]]*\]\([^)]*\)`)

	entityMap = map[string]string{
		"&amp;":  "&",
		"&lt;":   "<",
		"&gt;":   ">",
		"&quot;": `"`,
		"&apos;": "'",
		"&nbsp;": " ",
	}
)

func stripHTML(s string) string {
	s = reHTMLTag.ReplaceAllString(s, "")
	s = reHTMLEntity.ReplaceAllStringFunc(s, func(e string) string {
		if v, ok := entityMap[e]; ok {
			return v
		}
		return e
	})
	lines := strings.Split(s, "\n")
	var out []string
	blank := 0
	for _, l := range lines {
		if strings.TrimSpace(l) == "" {
			blank++
			if blank <= 1 {
				out = append(out, "")
			}
		} else {
			blank = 0
			out = append(out, l)
		}
	}
	return strings.Join(out, "\n")
}

func stripImages(s string) string {
	return reImage.ReplaceAllString(s, "")
}

// renderMarkdown renders markdown content with glamour and returns the lines.
func renderMarkdown(content string, width int) []string {
	if width < 1 {
		return strings.Split(content, "\n")
	}

	content = stripHTML(content)
	content = stripImages(content)

	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return strings.Split(content, "\n")
	}
	rendered, err := r.Render(content)
	if err != nil {
		return strings.Split(content, "\n")
	}
	rendered = strings.Trim(rendered, "\n")

	rawLines := strings.Split(rendered, "\n")
	var result []string
	for _, line := range rawLines {
		if xansi.StringWidth(line) > width {
			line = xansi.Truncate(line, width, "")
		}
		result = append(result, line)
	}
	return result
}
