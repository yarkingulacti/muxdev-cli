//go:build !windows

package portkill

import (
	"os"
	"runtime"
	"testing"
)

func TestFilterEmpty(t *testing.T) {
	got := filterEmpty([]string{"node", "", "server.js", ""})
	if len(got) != 2 || got[0] != "node" || got[1] != "server.js" {
		t.Fatalf("filterEmpty() = %#v", got)
	}
}

func TestReadCmdlineJoinsArgs(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("/proc cmdline is Linux-only")
	}
	cmd, err := readCmdline(os.Getpid())
	if err != nil {
		t.Fatalf("readCmdline: %v", err)
	}
	if cmd == "" {
		t.Fatal("expected non-empty cmdline")
	}
}
