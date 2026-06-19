package platform

import (
	"fmt"
	"os"
)

// SessionsDir returns the platform-specific directory for persisted runtime logs.
//
// Linux: $XDG_STATE_HOME/muxdev/sessions or ~/.local/state/muxdev/sessions
// macOS: ~/Library/Application Support/muxdev/sessions
// Windows: %LOCALAPPDATA%\muxdev\sessions
func SessionsDir() (string, error) {
	dir, err := sessionsDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create sessions dir: %w", err)
	}
	return dir, nil
}
