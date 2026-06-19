//go:build !windows

package update

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"
)

func applyReplace(target, source string) error {
	info, err := os.Stat(source)
	if err != nil {
		return err
	}
	mode := info.Mode() | 0o111

	if err := os.Chmod(source, mode); err != nil {
		return fmt.Errorf("chmod new binary: %w", err)
	}

	if err := os.Rename(source, target); err == nil {
		return nil
	} else if !sameDeviceReplaceError(err) {
		return fmt.Errorf("replace binary: %w", err)
	}

	data, err := os.ReadFile(source)
	if err != nil {
		return fmt.Errorf("read new binary: %w", err)
	}
	if err := os.WriteFile(target, data, mode); err != nil {
		return fmt.Errorf("replace binary: %w", err)
	}
	return nil
}

func sameDeviceReplaceError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "cross-device") ||
		strings.Contains(err.Error(), "cross device") ||
		errors.Is(err, syscall.EXDEV)
}
