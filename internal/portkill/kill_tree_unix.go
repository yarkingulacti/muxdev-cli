//go:build !windows

package portkill

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func expandKillTargets(pids []int) []int {
	seen := make(map[int]bool, len(pids)*2)
	out := make([]int, 0, len(pids)*2)

	add := func(pid int) {
		if pid <= 1 || seen[pid] {
			return
		}
		seen[pid] = true
		out = append(out, pid)
	}

	for _, pid := range pids {
		add(pid)
		for _, child := range childPIDs(pid) {
			add(child)
		}
		for cur := pid; cur > 1; {
			parent, err := parentPID(cur)
			if err != nil || parent <= 1 {
				break
			}
			add(parent)
			if shouldStopParentWalk(parent) {
				break
			}
			cur = parent
		}
	}

	return out
}

func shouldStopParentWalk(pid int) bool {
	cmd, err := readCmdline(pid)
	if err != nil {
		return false
	}
	lower := strings.ToLower(cmd)
	if strings.Contains(lower, "muxdev") {
		return true
	}
	return strings.HasPrefix(lower, "ss ") || strings.HasPrefix(lower, "lsof ")
}

func parentPID(pid int) (int, error) {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/status", pid))
	if err != nil {
		return 0, err
	}
	for _, line := range strings.Split(string(data), "\n") {
		if !strings.HasPrefix(line, "PPid:") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			break
		}
		return strconv.Atoi(fields[1])
	}
	return 0, fmt.Errorf("ppid not found for %d", pid)
}

func childPIDs(pid int) []int {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil
	}

	var children []int
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		child, err := strconv.Atoi(entry.Name())
		if err != nil || child <= 1 {
			continue
		}
		parent, err := parentPID(child)
		if err != nil || parent != pid {
			continue
		}
		children = append(children, child)
	}
	return children
}

func terminatePID(pid int) {
	signalProcessGroup(pid, syscall.SIGTERM)
}

func killPID(pid int) {
	signalProcessGroup(pid, syscall.SIGKILL)
}

func signalProcessGroup(pid int, sig syscall.Signal) {
	pgid, err := syscall.Getpgid(pid)
	if err == nil && pgid > 0 {
		_ = syscall.Kill(-pgid, sig)
	}
	proc, err := os.FindProcess(pid)
	if err == nil {
		_ = proc.Signal(sig)
	}
}

func fuserKillPort(port int) error {
	if _, err := exec.LookPath("fuser"); err != nil {
		return err
	}
	return exec.Command("fuser", "-k", fmt.Sprintf("%d/tcp", port)).Run()
}

func killPortWithRetry(port int) (int, error) {
	killed := 0

	for round := 0; round < 6; round++ {
		pids, err := PIDsOnPort(port)
		if err != nil {
			return killed, err
		}
		if len(pids) == 0 {
			return killed, nil
		}

		targets := expandKillTargets(pids)
		for _, pid := range targets {
			terminatePID(pid)
			killed++
		}

		time.Sleep(time.Duration(120*(round+1)) * time.Millisecond)

		for _, pid := range targets {
			if processAlive(pid) {
				killPID(pid)
			}
		}
	}

	remaining, _ := PIDsOnPort(port)
	if len(remaining) > 0 {
		_ = fuserKillPort(port)
		time.Sleep(300 * time.Millisecond)
	}

	remaining, err := PIDsOnPort(port)
	if err != nil {
		return killed, err
	}
	if len(remaining) == 0 {
		return killed, nil
	}

	proc, _ := ProcessOnPort(port)
	if proc.PID > 0 {
		cmd := proc.Command
		if cmd == "" {
			cmd = "unknown command"
		}
		return killed, fmt.Errorf("port %d still in use by pid %d (%s)", port, proc.PID, cmd)
	}
	return killed, fmt.Errorf("port %d still in use", port)
}
