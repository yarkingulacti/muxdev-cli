package update_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yarkingulacti/muxdev-cli/internal/update"
)

func TestVerifyChecksum(t *testing.T) {
	checksums := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855  muxdev_0.1.0_linux_amd64.tar.gz\n"
	dir := t.TempDir()
	file := filepath.Join(dir, "archive.tar.gz")
	if err := os.WriteFile(file, []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := update.VerifyChecksum(checksums, "muxdev_0.1.0_linux_amd64.tar.gz", file); err != nil {
		t.Fatalf("VerifyChecksum() error = %v", err)
	}
}

func TestPlatformAssetName(t *testing.T) {
	if got := update.PlatformAssetName("v0.1.0", "linux", "amd64"); got != "muxdev_0.1.0_linux_amd64.tar.gz" {
		t.Fatalf("linux asset = %q", got)
	}
	if got := update.PlatformAssetName("v0.1.0", "windows", "arm64"); got != "muxdev_0.1.0_windows_arm64.zip" {
		t.Fatalf("windows asset = %q", got)
	}
}

func TestDetectHomebrew(t *testing.T) {
	method := update.Detect("/opt/homebrew/Cellar/muxdev/0.1.0/bin/muxdev", "direct")
	if method != update.MethodHomebrew {
		t.Fatalf("method = %q, want homebrew", method)
	}
}

func TestDetectDirect(t *testing.T) {
	method := update.Detect("/home/user/.local/bin/muxdev", "direct")
	if method != update.MethodDirect {
		t.Fatalf("method = %q, want direct", method)
	}
}

func TestUpgradeHint(t *testing.T) {
	hint := update.MethodHomebrew.UpgradeHint("v0.2.0")
	if hint != "brew upgrade muxdev" {
		t.Fatalf("hint = %q", hint)
	}
}
