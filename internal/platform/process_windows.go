//go:build windows

package platform

import (
	"os/exec"
	"strconv"
	"time"
)

// TerminateProcessGroup sends a graceful shutdown to the service process tree.
func TerminateProcessGroup(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	_ = exec.Command("taskkill", "/T", "/PID", strconv.Itoa(cmd.Process.Pid)).Run()
}

// KillProcessGroup force-kills the service process tree.
func KillProcessGroup(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	_ = exec.Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(cmd.Process.Pid)).Run()
}

// StopProcessGroup tries graceful taskkill, then force-kills the tree.
func StopProcessGroup(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	TerminateProcessGroup(cmd)
	time.Sleep(400 * time.Millisecond)
	KillProcessGroup(cmd)
}
