//go:build !windows

package platform

import (
	"os"
	"os/exec"
	"syscall"
)

func ConfigureCommand(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

func interruptSignals() []os.Signal {
	return []os.Signal{os.Interrupt, syscall.SIGTERM}
}

func ShellCommand() string {
	return "/bin/sh"
}

func ShellArgs(command string) []string {
	return []string{"-c", command}
}
