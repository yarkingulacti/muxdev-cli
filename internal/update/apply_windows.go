//go:build windows

package update

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func applyReplace(target, source string) error {
	if err := os.Rename(target, target+".old"); err != nil {
		if err := stagePendingUpdate(target, source); err != nil {
			return fmt.Errorf("windows update: binary in use; run 'muxdev update --apply-pending' after closing other sessions: %w", err)
		}
		return nil
	}

	if err := copyFile(source, target); err != nil {
		return err
	}
	_ = os.Remove(target + ".old")
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}

func stagePendingUpdate(target, source string) error {
	pendingDir, err := pendingDirPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(pendingDir, 0o755); err != nil {
		return err
	}
	staged := filepath.Join(pendingDir, "muxdev.exe.new")
	if err := copyFile(source, staged); err != nil {
		return err
	}
	return savePendingUpdate(PendingUpdate{Target: target, Source: staged})
}
