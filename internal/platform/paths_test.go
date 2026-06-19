package platform

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestSessionsDirUsesXDGStateHome(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("XDG_STATE_HOME session path applies on Linux")
	}
	base := t.TempDir()
	t.Setenv("XDG_STATE_HOME", base)

	dir, err := SessionsDir()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(base, "muxdev", "sessions")
	if dir != want {
		t.Fatalf("SessionsDir() = %q, want %q", dir, want)
	}
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("sessions dir not created: %v", err)
	}
}
