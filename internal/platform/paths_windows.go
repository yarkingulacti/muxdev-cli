//go:build windows

package platform

import (
	"os"
	"path/filepath"
)

func sessionsDir() (string, error) {
	base := os.Getenv("LOCALAPPDATA")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, "AppData", "Local")
	}
	return filepath.Join(base, "muxdev", "sessions"), nil
}
