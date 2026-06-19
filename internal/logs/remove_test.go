package logs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRemoveSessionsDeletesProjectLogs(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())

	projectA := filepath.Join(t.TempDir(), "a")
	projectB := filepath.Join(t.TempDir(), "b")
	for _, dir := range []string{projectA, projectB} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	for _, spec := range []struct {
		dir       string
		serviceID string
	}{
		{projectA, "a"},
		{projectB, "b"},
	} {
		w, err := StartSession(spec.dir, filepath.Join(spec.dir, "muxdev.yaml"), []string{spec.serviceID}, "sync")
		if err != nil {
			t.Fatal(err)
		}
		if err := w.Finish(nil); err != nil {
			t.Fatal(err)
		}
	}

	removed, err := RemoveSessions(projectA)
	if err != nil {
		t.Fatalf("RemoveSessions() error = %v", err)
	}
	if removed != 1 {
		t.Fatalf("removed = %d, want 1", removed)
	}

	left, err := ListSessions(projectA)
	if err != nil {
		t.Fatal(err)
	}
	if len(left) != 0 {
		t.Fatalf("project A sessions = %d, want 0", len(left))
	}

	other, err := ListSessions(projectB)
	if err != nil {
		t.Fatal(err)
	}
	if len(other) != 1 {
		t.Fatalf("project B sessions = %d, want 1", len(other))
	}
}
