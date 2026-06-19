package config_test

import (
	"testing"

	"github.com/yarkingulacti/muxdev-cli/internal/config"
)

func TestParseRuntime(t *testing.T) {
	tests := []struct {
		in   string
		want config.Runtime
	}{
		{"", config.RuntimeSync},
		{"sync", config.RuntimeSync},
		{"sequential", config.RuntimeSync},
		{"async", config.RuntimeAsync},
		{"parallel", config.RuntimeAsync},
	}
	for _, tc := range tests {
		got, err := config.ParseRuntime(tc.in)
		if err != nil {
			t.Fatalf("ParseRuntime(%q) error = %v", tc.in, err)
		}
		if got != tc.want {
			t.Fatalf("ParseRuntime(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}

	if _, err := config.ParseRuntime("invalid"); err == nil {
		t.Fatal("ParseRuntime(invalid) expected error")
	}
}

func TestResolvedRuntime(t *testing.T) {
	cfg := &config.Config{Name: "App", Runtime: "async", Services: map[string]config.Service{
		"a": {Label: "A", Command: "true"},
	}}

	got, err := cfg.ResolvedRuntime("")
	if err != nil || got != config.RuntimeAsync {
		t.Fatalf("ResolvedRuntime() = %q, err %v", got, err)
	}

	got, err = cfg.ResolvedRuntime("sync")
	if err != nil || got != config.RuntimeSync {
		t.Fatalf("ResolvedRuntime(sync override) = %q, err %v", got, err)
	}
}

func TestValidateRuntime(t *testing.T) {
	cfg := &config.Config{
		Name:     "App",
		Runtime:  "nope",
		Services: map[string]config.Service{"a": {Label: "A", Command: "true"}},
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() expected runtime error")
	}
}
