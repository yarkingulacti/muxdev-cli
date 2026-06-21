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

	"golang.org/x/sys/unix"
)

// expandKillTargets returns the port-bound PIDs together with their full
// descendant tree.
//
// It deliberately never walks *up* the parent chain. Doing so used to drag
// ancestors such as the user's login shell or `systemd --user` into the kill
// set; combined with the process-group signalling below, that could tear down
// the entire login session when a target had been reparented out of muxdev's
// own tree (e.g. a leftover service from a previous run).
func expandKillTargets(pids []int) []int {
	seen := make(map[int]bool, len(pids)*2)
	out := make([]int, 0, len(pids)*2)

	add := func(pid int) bool {
		if pid <= 1 || seen[pid] {
			return false
		}
		seen[pid] = true
		out = append(out, pid)
		return true
	}

	childrenByParent := readProcessTree()

	var addSubtree func(pid int)
	addSubtree = func(pid int) {
		if !add(pid) {
			return
		}
		for _, child := range childrenByParent[pid] {
			addSubtree(child)
		}
	}

	for _, pid := range pids {
		addSubtree(pid)
	}

	return out
}

// readProcessTree scans /proc once and returns a parent->children map. On
// platforms without /proc (e.g. macOS) it returns nil, in which case callers
// fall back to signalling only the port-bound PIDs themselves.
func readProcessTree() map[int][]int {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil
	}

	tree := make(map[int][]int)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(entry.Name())
		if err != nil || pid <= 1 {
			continue
		}
		parent, err := parentPID(pid)
		if err != nil || parent <= 0 {
			continue
		}
		tree[parent] = append(tree[parent], pid)
	}
	return tree
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

func terminatePID(pid int) {
	signalProcessGroup(pid, syscall.SIGTERM)
}

func killPID(pid int) {
	signalProcessGroup(pid, syscall.SIGKILL)
}

// signalProcessGroup signals a target process. It escalates to the target's
// whole process group only when that group is safe to signal (see
// safeToSignalGroup); otherwise it signals just the single PID. This prevents
// a stray `kill(-pgid)` from reaching muxdev's own group or, worse, the login
// session's process group.
func signalProcessGroup(pid int, sig syscall.Signal) {
	if pgid, err := unix.Getpgid(pid); err == nil && safeToSignalGroup(pgid) {
		_ = syscall.Kill(-pgid, sig)
	}
	if proc, err := os.FindProcess(pid); err == nil {
		_ = proc.Signal(sig)
	}
}

// safeToSignalGroup reports whether `kill(-pgid)` may be used for the given
// process group without risking muxdev itself or the user's session.
func safeToSignalGroup(pgid int) bool {
	if pgid <= 1 {
		return false
	}
	// Never signal muxdev's own process group — that would kill the runner
	// (and, when muxdev shares the terminal's foreground group, the TUI).
	if own, err := unix.Getpgid(0); err == nil && pgid == own {
		return false
	}
	// A process group whose id is also a session id is a session leader's
	// group. Signalling it would tear down the whole login/terminal session.
	if sid, err := unix.Getsid(pgid); err != nil || sid == pgid {
		return false
	}
	return true
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
