package github

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// cacheDir はテスト時に上書き可能なキャッシュベースディレクトリ。
// 空の場合は os.UserCacheDir() を使う。
var cacheDir string

type repoCache struct {
	FetchedAt time.Time `json:"fetched_at"`
	Repos     []*Repo   `json:"repos"`
}

func cacheFilePath() (string, error) {
	dir := cacheDir
	if dir == "" {
		var err error
		dir, err = os.UserCacheDir()
		if err != nil {
			return "", err
		}
	}
	return filepath.Join(dir, "den-ghstarred", "repos.json"), nil
}

func loadCache() ([]*Repo, bool) {
	path, err := cacheFilePath()
	if err != nil {
		return nil, false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	var c repoCache
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, false
	}
	if time.Since(c.FetchedAt) > CacheTTL {
		return nil, false
	}
	return c.Repos, true
}

func clearCache() error {
	path, err := cacheFilePath()
	if err != nil {
		return err
	}
	return os.Remove(path)
}

func saveCache(repos []*Repo) {
	path, err := cacheFilePath()
	if err != nil {
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return
	}
	data, err := json.Marshal(repoCache{FetchedAt: time.Now(), Repos: repos})
	if err != nil {
		return
	}
	_ = os.WriteFile(path, data, 0o600)
}

// SaveCache writes repos to the on-disk cache. Call after modifying the starred list.
func SaveCache(repos []*Repo) {
	saveCache(repos)
}
