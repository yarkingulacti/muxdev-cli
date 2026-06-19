package tui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

// guessProjectName returns a human-readable project name from the work directory.
func guessProjectName(workDir string) string {
	if workDir == "" {
		return ""
	}
	base := filepath.Base(filepath.Clean(workDir))
	if base == "" || base == "." || base == string(filepath.Separator) {
		return ""
	}
	if strings.HasPrefix(base, ".") {
		return ""
	}
	return humanizeDirName(base)
}

func humanizeDirName(name string) string {
	name = strings.NewReplacer("-", " ", "_", " ").Replace(name)
	words := strings.Fields(name)
	for i, word := range words {
		if word == "" {
			continue
		}
		runes := []rune(strings.ToLower(word))
		runes[0] = unicode.ToUpper(runes[0])
		words[i] = string(runes)
	}
	return strings.Join(words, " ")
}

// guessDevCommand suggests a common dev start command based on project files.
func guessDevCommand(workDir string) string {
	if workDir == "" {
		return ""
	}
	if cmd := devCommandFromPackageJSON(workDir); cmd != "" {
		return cmd
	}
	if fileExists(filepath.Join(workDir, "go.mod")) {
		return "go run ."
	}
	if fileExists(filepath.Join(workDir, "Cargo.toml")) {
		return "cargo run"
	}
	if fileExists(filepath.Join(workDir, "mix.exs")) {
		return "mix phx.server"
	}
	return ""
}

func devCommandFromPackageJSON(workDir string) string {
	data, err := os.ReadFile(filepath.Join(workDir, "package.json"))
	if err != nil {
		return ""
	}
	var pkg struct {
		Scripts map[string]string `json:"scripts"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil || len(pkg.Scripts) == 0 {
		return ""
	}
	for _, script := range []string{"dev", "start", "serve"} {
		if strings.TrimSpace(pkg.Scripts[script]) != "" {
			return "npm run " + script
		}
	}
	return ""
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
