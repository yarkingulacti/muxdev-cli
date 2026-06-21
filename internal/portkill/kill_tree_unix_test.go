//go:build !windows

package portkill

import (
	"os"
	"os/exec"
	"runtime"
	"syscall"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

func contains(pids []int, want int) bool {
	for _, p := range pids {
		if p == want {
			return true
		}
	}
	return false
}

func startChild(t *testing.T, attr *syscall.SysProcAttr, args ...string) *exec.Cmd {
	t.Helper()
	cmd := exec.Command(args[0], args[1:]...)
	cmd.SysProcAttr = attr
	if err := cmd.Start(); err != nil {
		t.Fatalf("start %v: %v", args, err)
	}
	t.Cleanup(func() {
		_ = cmd.Process.Kill()
		_, _ = cmd.Process.Wait()
	})
	return cmd
}

// expandKillTargets must never drag in the caller's ancestors — only the
// port-bound PIDs and their descendants.
func TestExpandKillTargetsNoUpwardWalk(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("process-tree expansion relies on /proc (Linux-only)")
	}

	child := startChild(t, &syscall.SysProcAttr{Setpgid: true}, "sleep", "30")
	childPID := child.Process.Pid

	targets := expandKillTargets([]int{childPID})

	if !contains(targets, childPID) {
		t.Fatalf("expected targets to contain child %d, got %v", childPID, targets)
	}
	if contains(targets, os.Getpid()) {
		t.Fatalf("targets must not include the parent %d (no upward walk): %v", os.Getpid(), targets)
	}
}

// Descendants of a target must be collected, including grandchildren.
func TestExpandKillTargetsIncludesDescendants(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("process-tree expansion relies on /proc (Linux-only)")
	}

	// `sh` stays alive and forks `sleep` as a child process.
	parent := startChild(t, &syscall.SysProcAttr{Setpgid: true}, "sh", "-c", "sleep 30 & wait")
	parentPID := parent.Process.Pid

	// Wait for the grandchild to appear in the process tree.
	var targets []int
	for i := 0; i < 40; i++ {
		targets = expandKillTargets([]int{parentPID})
		if len(targets) >= 2 {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	if !contains(targets, parentPID) {
		t.Fatalf("expected targets to contain parent %d, got %v", parentPID, targets)
	}
	if len(targets) < 2 {
		t.Fatalf("expected at least one descendant of %d, got %v", parentPID, targets)
	}
}

func TestSafeToSignalGroupRejectsUnsafeGroups(t *testing.T) {
	if pgid, err := unix.Getpgid(0); err == nil {
		if safeToSignalGroup(pgid) {
			t.Errorf("muxdev's own process group %d must not be signal-safe", pgid)
		}
	}
	if safeToSignalGroup(0) || safeToSignalGroup(1) {
		t.Errorf("pgid 0/1 must not be signal-safe")
	}
}

// A session leader's group (pgid == sid) must be rejected so that signalling
// can never tear down the login/terminal session.
func TestSafeToSignalGroupRejectsSessionGroup(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("relies on /proc-backed session lookup")
	}

	leader := startChild(t, &syscall.SysProcAttr{Setsid: true}, "sleep", "30")
	pgid, err := unix.Getpgid(leader.Process.Pid)
	if err != nil {
		t.Fatalf("getpgid: %v", err)
	}
	if safeToSignalGroup(pgid) {
		t.Errorf("session leader group %d must not be signal-safe", pgid)
	}
}

// A dedicated sub-group within muxdev's own session (how spawned services run)
// is safe to signal as a group.
func TestSafeToSignalGroupAllowsServiceGroup(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("relies on /proc-backed session lookup")
	}

	svc := startChild(t, &syscall.SysProcAttr{Setpgid: true}, "sleep", "30")
	pgid, err := unix.Getpgid(svc.Process.Pid)
	if err != nil {
		t.Fatalf("getpgid: %v", err)
	}
	if !safeToSignalGroup(pgid) {
		t.Errorf("dedicated service group %d should be signal-safe", pgid)
	}
}
