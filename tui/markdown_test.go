package tui

import (
	"strings"
	"testing"
)

func TestRenderMarkdown_StripImages(t *testing.T) {
	md := "# Title\n\n![alt text](https://example.com/img.png)\n\nSome text after."
	lines := renderMarkdown(md, 80)
	joined := strings.Join(lines, "\n")
	if strings.Contains(joined, "example.com") {
		t.Errorf("renderMarkdown should strip images, got: %q", joined)
	}
	if !strings.Contains(joined, "Title") {
		t.Errorf("renderMarkdown should keep text, got: %q", joined)
	}
}

func TestRenderMarkdown_NoImages(t *testing.T) {
	md := "# Hello\n\nJust some text."
	lines := renderMarkdown(md, 80)
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "Hello") {
		t.Errorf("renderMarkdown output missing 'Hello': %q", joined)
	}
}

func TestStripImages(t *testing.T) {
	input := "Before ![alt](https://example.com/img.png) after"
	got := stripImages(input)
	if strings.Contains(got, "![") {
		t.Errorf("stripImages should remove image syntax, got: %q", got)
	}
	if !strings.Contains(got, "Before") || !strings.Contains(got, "after") {
		t.Errorf("stripImages should keep surrounding text, got: %q", got)
	}
}
