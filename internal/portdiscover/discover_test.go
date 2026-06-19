package portdiscover_test

import (
	"testing"

	"github.com/yarkingulacti/muxdev-cli/internal/portdiscover"
)

func TestExtractFromLines(t *testing.T) {
	tests := []struct {
		name string
		line string
		want string
	}{
		{"vite local", "  ➜  Local:   http://localhost:5173/", "5173"},
		{"localhost", "Server listening at http://localhost:4000", "4000"},
		{"127", "ready on http://127.0.0.1:3000", "3000"},
		{"port label", "Port: 8080", "8080"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := portdiscover.ParseLine(tt.line)
			if got != tt.want {
				t.Fatalf("ParseLine() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseLineEmpty(t *testing.T) {
	got := portdiscover.ParseLine("nothing useful here")
	if got != "" {
		t.Fatalf("ParseLine() = %q, want empty", got)
	}
}

func TestDiscoverCommand(t *testing.T) {
	got, err := portdiscover.Discover(t.Context(), ".", `printf '%s\n' 'Local: http://localhost:4321/'`)
	if err != nil {
		t.Fatalf("Discover() error = %v", err)
	}
	if got != "4321" {
		t.Fatalf("Discover() = %q, want 4321", got)
	}
}
