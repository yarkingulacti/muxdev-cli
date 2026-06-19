//go:build !windows

package platform

import (
	"os/exec"
	"syscall"
	"time"
)

// TerminateProcessGroup sends SIGTERM to the service process group.
func TerminateProcessGroup(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
}

// KillProcessGroup force-kills the service process group.
func KillProcessGroup(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
}

// StopProcessGroup tries SIGTERM, then SIGKILL after a short grace period.
func StopProcessGroup(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	TerminateProcessGroup(cmd)
	time.Sleep(400 * time.Millisecond)
	KillProcessGroup(cmd)
}
