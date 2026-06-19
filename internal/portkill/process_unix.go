//go:build !windows

package portkill

import (
	"fmt"
	"os"
	"strings"
)

// ProcessOnPort returns the first process listening on the given TCP port.
func ProcessOnPort(port int) (Process, error) {
	pids, err := PIDsOnPort(port)
	if err != nil {
		return Process{}, err
	}
	if len(pids) == 0 {
		return Process{}, fmt.Errorf("no process found on port %d", port)
	}
	cmd, err := readCmdline(pids[0])
	if err != nil {
		return Process{PID: pids[0], Command: ""}, nil
	}
	return Process{PID: pids[0], Command: cmd}, nil
}

func readCmdline(pid int) (string, error) {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid))
	if err != nil {
		return "", err
	}
	parts := strings.Split(strings.TrimRight(string(data), "\x00"), "\x00")
	parts = filterEmpty(parts)
	if len(parts) == 0 {
		return "", fmt.Errorf("empty cmdline for pid %d", pid)
	}
	return strings.Join(parts, " "), nil
}

func filterEmpty(parts []string) []string {
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}
