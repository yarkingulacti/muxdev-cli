//go:build windows

package platform

import (
	"os"
	"os/exec"
	"syscall"
)

func ConfigureCommand(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
}

func interruptSignals() []os.Signal {
	return []os.Signal{os.Interrupt}
}

func ShellCommand() string {
	return "cmd.exe"
}

func ShellArgs(command string) []string {
	return []string{"/C", command}
}
