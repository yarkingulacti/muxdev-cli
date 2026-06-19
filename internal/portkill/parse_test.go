package portkill_test

import (
	"testing"

	"github.com/yarkingulacti/muxdev-cli/internal/portkill"
)

func TestParseConflictEADDRINUSE(t *testing.T) {
	line := `Error: listen EADDRINUSE: address already in use :::4000 +1ms`
	c, ok := portkill.ParseConflict(line)
	if !ok {
		t.Fatal("expected conflict")
	}
	if c.Port != 4000 || !c.Fatal {
		t.Fatalf("conflict = %+v", c)
	}
}

func TestParseConflictPortInUse(t *testing.T) {
	line := "Port 5173 is in use, trying another one..."
	c, ok := portkill.ParseConflict(line)
	if !ok {
		t.Fatal("expected conflict")
	}
	if c.Port != 5173 || c.Fatal {
		t.Fatalf("conflict = %+v", c)
	}
}

func TestParseConflictPortAlreadyInUse(t *testing.T) {
	line := "Error: Port 5173 is already in use"
	c, ok := portkill.ParseConflict(line)
	if !ok {
		t.Fatal("expected conflict")
	}
	if c.Port != 5173 || !c.Fatal {
		t.Fatalf("conflict = %+v", c)
	}
}

func TestParseConflictNone(t *testing.T) {
	if _, ok := portkill.ParseConflict("all good"); ok {
		t.Fatal("expected no conflict")
	}
}
