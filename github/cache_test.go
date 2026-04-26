package github

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCacheRoundTrip(t *testing.T) {
	cacheDir = t.TempDir()
	t.Cleanup(func() { cacheDir = "" })

	repos := []*Repo{
		{Owner: "charmbracelet", Name: "bubbletea", FullName: "charmbracelet/bubbletea", StarCount: 100},
		{Owner: "golang", Name: "go", FullName: "golang/go", StarCount: 200},
	}

	SaveCache(repos)

	loaded, ok := loadCache()
	if !ok {
		t.Fatal("loadCache returned false; expected cache hit")
	}
	if len(loaded) != len(repos) {
		t.Fatalf("got %d repos, want %d", len(loaded), len(repos))
	}
	if loaded[0].FullName != "charmbracelet/bubbletea" {
		t.Errorf("got FullName %q, want %q", loaded[0].FullName, "charmbracelet/bubbletea")
	}
	if loaded[1].StarCount != 200 {
		t.Errorf("got StarCount %d, want 200", loaded[1].StarCount)
	}
}

func TestClearCache(t *testing.T) {
	cacheDir = t.TempDir()
	t.Cleanup(func() { cacheDir = "" })

	repos := []*Repo{{FullName: "foo/bar"}}
	SaveCache(repos)

	if err := clearCache(); err != nil {
		t.Fatalf("clearCache returned error: %v", err)
	}

	_, ok := loadCache()
	if ok {
		t.Error("loadCache returned true after clearCache; expected cache miss")
	}
}

func TestClearCacheNotExist(t *testing.T) {
	cacheDir = t.TempDir()
	t.Cleanup(func() { cacheDir = "" })

	// ファイルがない状態で clearCache を呼ぶと os.IsNotExist エラーを返す
	err := clearCache()
	if !os.IsNotExist(err) {
		t.Errorf("expected IsNotExist error, got: %v", err)
	}
}

func TestCacheExpiry(t *testing.T) {
	cacheDir = t.TempDir()
	t.Cleanup(func() { cacheDir = "" })

	// キャッシュファイルを直接作成して FetchedAt を2時間前に設定
	path, err := cacheFilePath()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}

	expired := repoCache{
		FetchedAt: time.Now().Add(-2 * time.Hour), // CacheTTL (1h) より古い
		Repos:     []*Repo{{FullName: "foo/bar"}},
	}
	data, err := json.Marshal(expired)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}

	_, ok := loadCache()
	if ok {
		t.Error("loadCache returned true for expired cache; expected false")
	}
}
