package tui

import (
	"strings"
	"testing"

	gh "github.com/ssshhhooota/den-ghstarred/github"
)

func TestBuildVisualItems_GroupsByOwner(t *testing.T) {
	repos := []*gh.Repo{
		{Owner: "charmbracelet", Name: "bubbletea", FullName: "charmbracelet/bubbletea"},
		{Owner: "charmbracelet", Name: "lipgloss", FullName: "charmbracelet/lipgloss"},
		{Owner: "golang", Name: "go", FullName: "golang/go"},
	}

	items := buildVisualItems(repos)

	// 期待: [owner:charmbracelet, repo:0, repo:1, owner:golang, repo:2]
	if len(items) != 5 {
		t.Fatalf("got %d items, want 5", len(items))
	}
	if !items[0].isOwner || items[0].owner != "charmbracelet" {
		t.Errorf("items[0]: got isOwner=%v owner=%q, want isOwner=true owner=%q", items[0].isOwner, items[0].owner, "charmbracelet")
	}
	if items[1].isOwner || items[1].repoIdx != 0 {
		t.Errorf("items[1]: got isOwner=%v repoIdx=%d, want isOwner=false repoIdx=0", items[1].isOwner, items[1].repoIdx)
	}
	if items[2].isOwner || items[2].repoIdx != 1 {
		t.Errorf("items[2]: got isOwner=%v repoIdx=%d, want isOwner=false repoIdx=1", items[2].isOwner, items[2].repoIdx)
	}
	if !items[3].isOwner || items[3].owner != "golang" {
		t.Errorf("items[3]: got isOwner=%v owner=%q, want isOwner=true owner=%q", items[3].isOwner, items[3].owner, "golang")
	}
	if items[4].isOwner || items[4].repoIdx != 2 {
		t.Errorf("items[4]: got isOwner=%v repoIdx=%d, want isOwner=false repoIdx=2", items[4].isOwner, items[4].repoIdx)
	}
}

func TestBuildVisualItems_NonConsecutiveSameOwner(t *testing.T) {
	repos := []*gh.Repo{
		{Owner: "charmbracelet", Name: "bubbletea", FullName: "charmbracelet/bubbletea"},
		{Owner: "golang", Name: "go", FullName: "golang/go"},
		{Owner: "charmbracelet", Name: "lipgloss", FullName: "charmbracelet/lipgloss"},
	}

	items := buildVisualItems(repos)

	// charmbracelet のリポジトリは連続していなくても同じグループに入る
	// 期待: [owner:charmbracelet, repo:0, repo:2, owner:golang, repo:1]
	if len(items) != 5 {
		t.Fatalf("got %d items, want 5", len(items))
	}
	if !items[0].isOwner || items[0].owner != "charmbracelet" {
		t.Errorf("items[0] should be owner charmbracelet, got %+v", items[0])
	}
	if items[1].repoIdx != 0 {
		t.Errorf("items[1].repoIdx = %d, want 0", items[1].repoIdx)
	}
	if items[2].repoIdx != 2 {
		t.Errorf("items[2].repoIdx = %d, want 2", items[2].repoIdx)
	}
	if !items[3].isOwner || items[3].owner != "golang" {
		t.Errorf("items[3] should be owner golang, got %+v", items[3])
	}
	if items[4].repoIdx != 1 {
		t.Errorf("items[4].repoIdx = %d, want 1", items[4].repoIdx)
	}
}

func TestBuildVisualItems_Empty(t *testing.T) {
	items := buildVisualItems(nil)
	if len(items) != 0 {
		t.Errorf("got %d items, want 0", len(items))
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input  string
		maxLen int
		want   string
	}{
		{"hello", 10, "hello"},
		{"hello world", 5, "hell…"},
		{"hi", 2, "hi"},
		{"x", 1, "x"},
		{"abc", 3, "abc"},
		{"abcd", 3, "ab…"},
		{"xy", 1, "…"},
	}
	for _, tt := range tests {
		got := truncate(tt.input, tt.maxLen)
		if got != tt.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
		}
	}
}

func TestStripANSI(t *testing.T) {
	input := "\x1b[1;32mhello\x1b[0m world"
	want := "hello world"
	got := stripANSI(input)
	if got != want {
		t.Errorf("stripANSI(%q) = %q, want %q", input, got, want)
	}
}

func TestRenderRight_BasicLines(t *testing.T) {
	m := Model{
		readme: ReadmePanel{
			lines: []string{"line one", "line two", "line three"},
		},
	}
	result := renderRight(m, 80, 3)
	if !strings.Contains(result, "line one") {
		t.Errorf("renderRight should contain 'line one', got: %q", result)
	}
}
