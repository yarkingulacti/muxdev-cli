package update

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type CacheEntry struct {
	CheckedAt       time.Time `json:"checked_at"`
	Current         string    `json:"current"`
	Latest          string    `json:"latest"`
	UpdateAvailable bool      `json:"update_available"`
}

func cacheFilePath() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		home, herr := os.UserHomeDir()
		if herr != nil {
			return "", herr
		}
		base = filepath.Join(home, ".cache")
	}
	return filepath.Join(base, "muxdev", "update.json"), nil
}

func ReadCache() (*CacheEntry, error) {
	path, err := cacheFilePath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, err
	}
	return &entry, nil
}

func WriteCache(entry CacheEntry) error {
	path, err := cacheFilePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func ShouldCheckOnStart(interval time.Duration) bool {
	if os.Getenv("MUXDEV_NO_UPDATE_CHECK") != "" {
		return false
	}
	entry, err := ReadCache()
	if err != nil || entry == nil {
		return true
	}
	return time.Since(entry.CheckedAt) >= interval
}

func StartupHint() string {
	entry, err := ReadCache()
	if err != nil || entry == nil || !entry.UpdateAvailable {
		return ""
	}
	return "Update available: " + entry.Latest + " — run: muxdev update"
}
