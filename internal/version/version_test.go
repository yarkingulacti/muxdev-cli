package version_test

import (
	"strings"
	"testing"

	"github.com/yarkingulacti/muxdev-cli/internal/version"
)

func TestStringDev(t *testing.T) {
	version.Version = "dev"
	if got := version.String(); !strings.Contains(got, "dev") {
		t.Fatalf("String() = %q", got)
	}
}

func TestJSON(t *testing.T) {
	version.Version = "0.1.0"
	out, err := version.JSON()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "0.1.0") {
		t.Fatalf("JSON() = %q", out)
	}
}
