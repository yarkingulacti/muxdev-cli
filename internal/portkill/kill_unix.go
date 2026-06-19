//go:build !windows

package portkill

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

// KillPort terminates processes listening on the given TCP port.
func KillPort(port int) (int, error) {
	return killPortWithRetry(port)
}

func PIDsOnPort(port int) ([]int, error) {
	if pids, err := pidsViaLsof(port); err == nil {
		return pids, nil
	}
	return pidsViaSS(port)
}

func pidsViaLsof(port int) ([]int, error) {
	out, err := exec.Command("lsof", "-ti", fmt.Sprintf(":%d", port)).Output()
	if err != nil {
		return nil, err
	}
	return parsePIDLines(string(out))
}

func pidsViaSS(port int) ([]int, error) {
	out, err := exec.Command("ss", "-lptn", fmt.Sprintf("sport = :%d", port)).Output()
	if err != nil {
		return nil, fmt.Errorf("find port %d: need lsof or ss", port)
	}
	return parseSSOutput(string(out))
}

func parsePIDLines(raw string) ([]int, error) {
	var pids []int
	for _, line := range strings.Split(strings.TrimSpace(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		pid, err := strconv.Atoi(line)
		if err != nil {
			continue
		}
		pids = appendUnique(pids, pid)
	}
	return pids, nil
}

func parseSSOutput(raw string) ([]int, error) {
	var pids []int
	for _, line := range strings.Split(raw, "\n") {
		idx := strings.Index(line, "pid=")
		if idx < 0 {
			continue
		}
		rest := line[idx+4:]
		end := strings.IndexAny(rest, ",)")
		if end > 0 {
			rest = rest[:end]
		}
		pid, err := strconv.Atoi(strings.TrimSpace(rest))
		if err != nil {
			continue
		}
		pids = appendUnique(pids, pid)
	}
	if len(pids) == 0 {
		return nil, fmt.Errorf("no pid in ss output")
	}
	return pids, nil
}

func appendUnique(pids []int, pid int) []int {
	for _, existing := range pids {
		if existing == pid {
			return pids
		}
	}
	return append(pids, pid)
}

func processAlive(pid int) bool {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return proc.Signal(syscall.Signal(0)) == nil
}
