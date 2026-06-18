package update

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type PendingUpdate struct {
	Target string `json:"target"`
	Source string `json:"source"`
}

func pendingDirPath() (string, error) {
	base := os.Getenv("LOCALAPPDATA")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, "AppData", "Local")
	}
	return filepath.Join(base, "muxdev"), nil
}

func pendingFilePath() (string, error) {
	dir, err := pendingDirPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "update-pending.json"), nil
}

func savePendingUpdate(update PendingUpdate) error {
	path, err := pendingFilePath()
	if err != nil {
		return err
	}
	data, err := json.Marshal(update)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func loadPendingUpdate() (*PendingUpdate, error) {
	path, err := pendingFilePath()
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
	var update PendingUpdate
	if err := json.Unmarshal(data, &update); err != nil {
		return nil, err
	}
	return &update, nil
}

func clearPendingUpdate() error {
	path, err := pendingFilePath()
	if err != nil {
		return err
	}
	err = os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}
