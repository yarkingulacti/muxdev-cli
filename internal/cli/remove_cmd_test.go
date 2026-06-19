package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yarkingulacti/muxdev-cli/internal/config"
	"github.com/yarkingulacti/muxdev-cli/internal/logs"
)

func TestRunRemoveDeletesConfigAndSessions(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, config.DefaultFilename)
	if err := os.WriteFile(cfgPath, []byte("name: Test\nservices:\n  api:\n    label: API\n    command: true\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)

	writer, err := logs.StartSession(dir, cfgPath, []string{"api"}, "sync")
	if err != nil {
		t.Fatal(err)
	}
	if err := writer.Finish(nil); err != nil {
		t.Fatal(err)
	}

	if err := runRemove("", true); err != nil {
		t.Fatalf("runRemove() error = %v", err)
	}
	if config.Exists(cfgPath) {
		t.Fatal("config should be removed")
	}
	sessions, err := logs.ListSessions(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 0 {
		t.Fatalf("sessions = %d, want 0", len(sessions))
	}
}

func TestRunRemoveNothingFound(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	if err := runRemove("", true); err == nil {
		t.Fatal("runRemove() expected error, got nil")
	}
}

func TestResolveRemoveTargetsUsesParentConfig(t *testing.T) {
	root := t.TempDir()
	sub := filepath.Join(root, "apps", "ui")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	cfgPath := filepath.Join(root, config.DefaultFilename)
	if err := os.WriteFile(cfgPath, []byte("name: Test\nservices:\n  api:\n    label: API\n    command: true\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Chdir(sub)
	gotPath, workDir, err := resolveRemoveTargets("")
	if err != nil {
		t.Fatalf("resolveRemoveTargets() error = %v", err)
	}
	if gotPath != cfgPath {
		t.Fatalf("cfgPath = %q, want %q", gotPath, cfgPath)
	}
	if workDir != root {
		t.Fatalf("workDir = %q, want %q", workDir, root)
	}
}
