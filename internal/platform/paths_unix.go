//go:build !windows

package platform

import (
	"os"
	"path/filepath"
	"runtime"
)

func sessionsDir() (string, error) {
	if runtime.GOOS == "darwin" {
		base, err := os.UserConfigDir()
		if err != nil {
			home, herr := os.UserHomeDir()
			if herr != nil {
				return "", herr
			}
			base = filepath.Join(home, "Library", "Application Support")
		}
		return filepath.Join(base, "muxdev", "sessions"), nil
	}

	if state := os.Getenv("XDG_STATE_HOME"); state != "" {
		return filepath.Join(state, "muxdev", "sessions"), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "state", "muxdev", "sessions"), nil
}
