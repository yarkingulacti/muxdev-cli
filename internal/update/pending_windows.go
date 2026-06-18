//go:build windows

package update

import "fmt"

func ApplyPending() error {
	pending, err := loadPendingUpdate()
	if err != nil {
		return err
	}
	if pending == nil {
		return fmt.Errorf("no pending update")
	}
	if err := applyReplace(pending.Target, pending.Source); err != nil {
		return fmt.Errorf("apply pending update: %w", err)
	}
	return clearPendingUpdate()
}
