package portkill_test

import (
	"testing"

	"github.com/yarkingulacti/muxdev-cli/internal/portkill"
)

func TestParseConflictEADDRINUSE(t *testing.T) {
	line := `Error: listen EADDRINUSE: address already in use :::4000 +1ms`
	c, ok := portkill.ParseConflict(line, 0)
	if !ok {
		t.Fatal("expected conflict")
	}
	if c.Port != 4000 || !c.Fatal {
		t.Fatalf("conflict = %+v", c)
	}
}

func TestParseConflictPortInUse(t *testing.T) {
	line := "Port 5173 is in use, trying another one..."
	c, ok := portkill.ParseConflict(line, 0)
	if !ok {
		t.Fatal("expected conflict")
	}
	if c.Port != 5173 || c.Fatal {
		t.Fatalf("conflict = %+v", c)
	}
}

func TestParseConflictPortAlreadyInUse(t *testing.T) {
	line := "Error: Port 5173 is already in use"
	c, ok := portkill.ParseConflict(line, 0)
	if !ok {
		t.Fatal("expected conflict")
	}
	if c.Port != 5173 || !c.Fatal {
		t.Fatalf("conflict = %+v", c)
	}
}

func TestParseConflictNone(t *testing.T) {
	if _, ok := portkill.ParseConflict("all good", 0); ok {
		t.Fatal("expected no conflict")
	}
}

func TestParseConflictNextJSLine(t *testing.T) {
	line := "Error: listen EADDRINUSE: address already in use :::4000"
	c, ok := portkill.ParseConflict(line, 3131)
	if !ok {
		t.Fatal("expected conflict")
	}
	if c.Port != 4000 {
		t.Fatalf("conflict port = %d, want 4000 (hint must not override)", c.Port)
	}
}

func TestParseConflictNestedNextJSLine(t *testing.T) {
	line := "    at <unknown> (Error: listen EADDRINUSE: address already in use :::4000)"
	c, ok := portkill.ParseConflict(line, 3131)
	if !ok || c.Port != 4000 {
		t.Fatalf("conflict = %+v, ok=%v", c, ok)
	}
}

func TestParseConflictJSONPortField(t *testing.T) {
	line := "  port: 4000"
	c, ok := portkill.ParseConflict(line, 3131)
	if !ok || c.Port != 4000 {
		t.Fatalf("conflict = %+v, ok=%v", c, ok)
	}
}

func TestParseConflictUvicornErrno98(t *testing.T) {
	line := "ERROR:    [Errno 98] Address already in use"
	c, ok := portkill.ParseConflict(line, 5005)
	if !ok {
		t.Fatal("expected conflict")
	}
	if c.Port != 5005 || !c.Fatal {
		t.Fatalf("conflict = %+v", c)
	}
}
