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

	env := config.ListingEnv(dir, nil)
	got := config.ExpandEnv("${UI_PORT}", env)
	if got != "4000" {
		t.Fatalf("ExpandEnv() = %q, want 4000", got)
	}
}

func TestExpandEnvKeepsUnknownPlaceholder(t *testing.T) {
	env := config.ListingEnv(t.TempDir(), nil)
	got := config.ExpandEnv("${MISSING_PORT}", env)
	if got != "${MISSING_PORT}" {
		t.Fatalf("ExpandEnv() = %q, want ${MISSING_PORT}", got)
	}
}

func TestExpandEnvDotenvOverridesServiceEnv(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("UI_PORT=4000\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	env := config.ListingEnv(dir, map[string]string{"UI_PORT": "3001"})
	got := config.ExpandEnv("${UI_PORT}", env)
	if got != "4000" {
		t.Fatalf("ExpandEnv() = %q, want 4000", got)
	}
}

func TestExpandEnvFromEnvExampleFallback(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".env.example"), []byte("BACKEND_PORT=5005\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	env := config.ListingEnv(dir, nil)
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

	env := config.ListingEnv(dir, nil)
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

	env := config.ListingEnv(dir, nil)
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

func TestListingEnvInterpolatesDotenv(t *testing.T) {
	dir := t.TempDir()
	content := "POSTGRES_USER=postgres\nPOSTGRES_PASSWORD=secret\nPOSTGRES_PORT=5432\nPOSTGRES_DB=app\nDATABASE_URL=postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:${POSTGRES_PORT}/${POSTGRES_DB}\n"
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	env := config.ListingEnv(dir, nil)
	got := env["DATABASE_URL"]
	want := "postgresql://postgres:secret@localhost:5432/app"
	if got != want {
		t.Fatalf("DATABASE_URL = %q, want %q", got, want)
	}
}

func TestPortEnvPriorityOrder(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("UI_PORT=4000\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".env.local"), []byte("UI_PORT=4100\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".env.example"), []byte("UI_PORT=4200\n# UI_PORT=4300\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		Services: map[string]config.Service{
			"ui": {
				Port: "${UI_PORT}",
				Env:  map[string]string{"UI_PORT": "3001"},
			},
		},
	}
	svc := cfg.Services["ui"]

	t.Run("shell", func(t *testing.T) {
		t.Setenv("UI_PORT", "8080")
		got := config.ExpandServicePort(cfg, dir, svc)
		if got != "8080" {
			t.Fatalf("ExpandServicePort() = %q, want 8080", got)
		}
	})

	t.Run("env local", func(t *testing.T) {
		got := config.ExpandServicePort(cfg, dir, svc)
		if got != "4100" {
			t.Fatalf("ExpandServicePort() = %q, want 4100", got)
		}
	})

	t.Run("env", func(t *testing.T) {
		if err := os.Remove(filepath.Join(dir, ".env.local")); err != nil {
			t.Fatal(err)
		}
		got := config.ExpandServicePort(cfg, dir, svc)
		if got != "4000" {
			t.Fatalf("ExpandServicePort() = %q, want 4000", got)
		}
	})

	t.Run("env example", func(t *testing.T) {
		if err := os.Remove(filepath.Join(dir, ".env")); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, ".env.local"), []byte("UI_PORT=4100\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.Remove(filepath.Join(dir, ".env.local")); err != nil {
			t.Fatal(err)
		}
		got := config.ExpandServicePort(cfg, dir, svc)
		if got != "4200" {
			t.Fatalf("ExpandServicePort() = %q, want 4200", got)
		}
	})

	t.Run("commented env example", func(t *testing.T) {
		if err := os.WriteFile(filepath.Join(dir, ".env.example"), []byte("# UI_PORT=4300\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		got := config.ExpandServicePort(cfg, dir, svc)
		if got != "4300" {
			t.Fatalf("ExpandServicePort() = %q, want 4300", got)
		}
	})

	t.Run("service env", func(t *testing.T) {
		if err := os.Remove(filepath.Join(dir, ".env.example")); err != nil {
			t.Fatal(err)
		}
		got := config.ExpandServicePort(cfg, dir, svc)
		if got != "3001" {
			t.Fatalf("ExpandServicePort() = %q, want 3001", got)
		}
	})
}

func TestExpandServicePortFromDotenv(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("UI_PORT=4000\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		Services: map[string]config.Service{
			"ui": {Port: "${UI_PORT}"},
		},
	}

	got := config.ExpandServicePort(cfg, dir, cfg.Services["ui"])
	if got != "4000" {
		t.Fatalf("ExpandServicePort() = %q, want 4000", got)
	}
}

func TestResolveServicePortSource(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("UI_PORT=4000\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		Services: map[string]config.Service{
			"ui":     {Port: "${UI_PORT}"},
			"api":    {Port: "5005"},
			"worker": {Port: "${WORKER_PORT:-9000}"},
		},
	}

	res := config.ResolveServicePort(cfg, dir, cfg.Services["ui"])
	if res.Port != "4000" || res.Source != ".env" {
		t.Fatalf("ResolveServicePort(ui) = %+v, want port 4000 from .env", res)
	}

	res = config.ResolveServicePort(cfg, dir, cfg.Services["api"])
	if res.Port != "5005" || res.Source != "muxdev.yaml" {
		t.Fatalf("ResolveServicePort(api) = %+v, want port 5005 from muxdev.yaml", res)
	}

	res = config.ResolveServicePort(cfg, dir, cfg.Services["worker"])
	if res.Port != "9000" || res.Source != "default" {
		t.Fatalf("ResolveServicePort(worker) = %+v, want port 9000 from default", res)
	}
}
