package logs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStartSessionWriteAndList(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())

	workDir := filepath.Join(t.TempDir(), "project")
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(workDir, "muxdev.yaml")

	writer, err := StartSession(workDir, configPath, []string{"backend", "ui"}, "sync")
	if err != nil {
		t.Fatal(err)
	}
	if err := writer.Append("Backend", "started"); err != nil {
		t.Fatal(err)
	}
	if err := writer.Finish(nil); err != nil {
		t.Fatal(err)
	}

	sessions, err := ListSessions(workDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 1 {
		t.Fatalf("sessions = %d, want 1", len(sessions))
	}
	if len(sessions[0].Meta.ServiceIDs) != 2 {
		t.Fatalf("service ids = %v", sessions[0].Meta.ServiceIDs)
	}

	content, err := ReadLog(sessions[0].Dir)
	if err != nil {
		t.Fatal(err)
	}
	if content != "[Backend] started\n" {
		t.Fatalf("log content = %q", content)
	}
}

func TestListSessionsFiltersOtherProjects(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())

	projectA := filepath.Join(t.TempDir(), "a")
	projectB := filepath.Join(t.TempDir(), "b")
	for _, dir := range []string{projectA, projectB} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	wA, err := StartSession(projectA, filepath.Join(projectA, "muxdev.yaml"), []string{"a"}, "sync")
	if err != nil {
		t.Fatal(err)
	}
	if err := wA.Finish(nil); err != nil {
		t.Fatal(err)
	}

	wB, err := StartSession(projectB, filepath.Join(projectB, "muxdev.yaml"), []string{"b"}, "sync")
	if err != nil {
		t.Fatal(err)
	}
	if err := wB.Finish(nil); err != nil {
		t.Fatal(err)
	}

	sessions, err := ListSessions(projectA)
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 1 || sessions[0].Meta.ServiceIDs[0] != "a" {
		t.Fatalf("filtered sessions = %+v", sessions)
	}
}
