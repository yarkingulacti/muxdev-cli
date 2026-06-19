package tui

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGuessProjectName(t *testing.T) {
	tests := []struct {
		dir  string
		want string
	}{
		{"/home/user/my-app", "My App"},
		{"/home/user/muxdev_cli", "Muxdev Cli"},
		{"/home/user/.hidden", ""},
		{"", ""},
	}
	for _, tt := range tests {
		if got := guessProjectName(tt.dir); got != tt.want {
			t.Errorf("guessProjectName(%q) = %q, want %q", tt.dir, got, tt.want)
		}
	}
}

func TestGuessDevCommandFromPackageJSON(t *testing.T) {
	dir := t.TempDir()
	pkg := `{"scripts":{"build":"vite build","dev":"vite"}}`
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkg), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := guessDevCommand(dir); got != "npm run dev" {
		t.Fatalf("guessDevCommand() = %q, want npm run dev", got)
	}
}

func TestGuessDevCommandFromGoMod(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/app\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := guessDevCommand(dir); got != "go run ." {
		t.Fatalf("guessDevCommand() = %q, want go run .", got)
	}
}
