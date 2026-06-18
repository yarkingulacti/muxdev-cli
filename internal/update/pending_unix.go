//go:build !windows

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
	return fmt.Errorf("pending updates are only used on Windows")
}
