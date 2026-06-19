package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yarkingulacti/muxdev-cli/internal/config"
)

func TestExpandEnvFromDotenv(t *testing.T) {
	dir := t.TempDir()
	content := "UI_PORT=4000\nBACKEND_PORT=5005\n"
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	env := config.ListingEnv(dir, nil, nil)
	got := config.ExpandEnv("${UI_PORT}", env)
	if got != "4000" {
		t.Fatalf("ExpandEnv() = %q, want 4000", got)
	}
}

func TestExpandEnvKeepsUnknownPlaceholder(t *testing.T) {
	env := config.ListingEnv(t.TempDir(), nil, nil)
	got := config.ExpandEnv("${MISSING_PORT}", env)
	if got != "${MISSING_PORT}" {
		t.Fatalf("ExpandEnv() = %q, want ${MISSING_PORT}", got)
	}
}

func TestExpandEnvServiceEnvOverridesDotenv(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("UI_PORT=4000\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	env := config.ListingEnv(dir, nil, map[string]string{"UI_PORT": "3001"})
	got := config.ExpandEnv("${UI_PORT}", env)
	if got != "3001" {
		t.Fatalf("ExpandEnv() = %q, want 3001", got)
	}
}

func TestExpandEnvFromEnvExampleFallback(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".env.example"), []byte("BACKEND_PORT=5005\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	env := config.ListingEnv(dir, nil, nil)
	got := config.ExpandEnv("${BACKEND_PORT}", env)
	if got != "5005" {
		t.Fatalf("ExpandEnv() = %q, want 5005", got)
	}
}

func TestExpandEnvFromCommentedEnvExample(t *testing.T) {
	dir := t.TempDir()
	content := "BACKEND_PORT=5005\n# OLLAMA_PORT=11435\n"
	if err := os.WriteFile(filepath.Join(dir, ".env.example"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	env := config.ListingEnv(dir, nil, nil)
	got := config.ExpandEnv("${OLLAMA_PORT}", env)
	if got != "11435" {
		t.Fatalf("ExpandEnv() = %q, want 11435", got)
	}
}

func TestExpandEnvOSOverridesDotenv(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("UI_PORT=4000\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("UI_PORT", "8080")

	env := config.ListingEnv(dir, nil, nil)
	got := config.ExpandEnv("${UI_PORT}", env)
	if got != "8080" {
		t.Fatalf("ExpandEnv() = %q, want 8080", got)
	}
}

func TestExpandEnvShellDefault(t *testing.T) {
	env := map[string]string{}
	got := config.ExpandEnv("${BACKEND_PORT:-5005}", env)
	if got != "5005" {
		t.Fatalf("ExpandEnv() = %q, want 5005", got)
	}
}

func TestListingEnvSourcesScript(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("BASE_PORT=9000\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	scriptDir := filepath.Join(dir, "scripts")
	if err := os.MkdirAll(scriptDir, 0o755); err != nil {
		t.Fatal(err)
	}
	script := `#!/usr/bin/env bash
set -a
source "$(dirname "${BASH_SOURCE[0]}")/../.env"
set +a
export UI_PORT="${BASE_PORT}"
`
	if err := os.WriteFile(filepath.Join(scriptDir, "muxdev-env.sh"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	env := config.ListingEnv(dir, []string{"scripts/muxdev-env.sh"}, nil)
	got := config.ExpandEnv("${UI_PORT}", env)
	if got != "9000" {
		t.Fatalf("ExpandEnv() = %q, want 9000", got)
	}
}

func TestListingEnvInterpolatesDotenvWithSource(t *testing.T) {
	dir := t.TempDir()
	content := "POSTGRES_USER=postgres\nPOSTGRES_PASSWORD=secret\nPOSTGRES_PORT=5432\nPOSTGRES_DB=app\nDATABASE_URL=postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:${POSTGRES_PORT}/${POSTGRES_DB}\n"
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	env := config.ListingEnv(dir, nil, nil)
	got := env["DATABASE_URL"]
	want := "postgresql://postgres:secret@localhost:5432/app"
	if got != want {
		t.Fatalf("DATABASE_URL = %q, want %q", got, want)
	}
}
