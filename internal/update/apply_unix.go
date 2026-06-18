//go:build !windows

package update

import (
	"fmt"
	"os"
)

func applyReplace(target, source string) error {
	info, err := os.Stat(source)
	if err != nil {
		return err
	}

	if err := os.Chmod(source, info.Mode()|0o111); err != nil {
		return fmt.Errorf("chmod new binary: %w", err)
	}

	if err := os.Rename(source, target); err != nil {
		return fmt.Errorf("replace binary: %w", err)
	}
	return nil
}
