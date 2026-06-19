package config_test

import (
	"path/filepath"
	"testing"

	"github.com/yarkingulacti/muxdev-cli/internal/config"
)

func TestLoadValidConfig(t *testing.T) {
	path := filepath.Join("..", "..", "testdata", "muxdev.yaml")
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Name != "My App" {
		t.Fatalf("Name = %q, want %q", cfg.Name, "My App")
	}

	ids, err := cfg.SortedServiceIDs()
	if err != nil {
		t.Fatalf("SortedServiceIDs() error = %v", err)
	}
	if len(ids) != 2 {
		t.Fatalf("len(ids) = %d, want 2", len(ids))
	}
	if ids[0] != "backend" {
		t.Fatalf("ids[0] = %q, want backend", ids[0])
	}
	if ids[1] != "ui" {
		t.Fatalf("ids[1] = %q, want ui", ids[1])
	}
}

func TestResolveServicesWithDependencies(t *testing.T) {
	path := filepath.Join("..", "..", "testdata", "muxdev.yaml")
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	ids, err := cfg.ResolveServices([]string{"ui"})
	if err != nil {
		t.Fatalf("ResolveServices() error = %v", err)
	}
	if len(ids) != 2 {
		t.Fatalf("len(ids) = %d, want 2", len(ids))
	}
	if ids[0] != "backend" || ids[1] != "ui" {
		t.Fatalf("ids = %v, want [backend ui]", ids)
	}
}

func TestValidateCycle(t *testing.T) {
	cfg := &config.Config{
		Name: "Cycle",
		Services: map[string]config.Service{
			"a": {Label: "A", Command: "true", DependsOn: []string{"b"}},
			"b": {Label: "B", Command: "true", DependsOn: []string{"a"}},
		},
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() expected cycle error, got nil")
	}
}

func TestOrderForStartDependencyOrder(t *testing.T) {
	services := map[string]config.Service{
		"backend": {Label: "Backend", Command: "true"},
		"ollama":  {Label: "Ollama", Command: "true"},
		"ui":      {Label: "UI", Command: "true", DependsOn: []string{"backend"}},
	}
	got := config.OrderForStart([]string{"ui", "ollama", "backend"}, services)
	want := []string{"backend", "ollama", "ui"}
	if len(got) != len(want) {
		t.Fatalf("OrderForStart() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("OrderForStart() = %v, want %v", got, want)
		}
	}
}

func TestOrderForStartCycleFallback(t *testing.T) {
	services := map[string]config.Service{
		"a": {Label: "A", Command: "true", DependsOn: []string{"b"}},
		"b": {Label: "B", Command: "true", DependsOn: []string{"a"}},
	}
	got := config.OrderForStart([]string{"b", "a"}, services)
	want := []string{"a", "b"}
	if len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("OrderForStart() = %v, want %v", got, want)
	}
}
