package update

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type InstallMethod string

const (
	MethodDirect   InstallMethod = "direct"
	MethodHomebrew InstallMethod = "homebrew"
	MethodScoop    InstallMethod = "scoop"
	MethodWinget   InstallMethod = "winget"
	MethodGoInstall InstallMethod = "go"
	MethodDev      InstallMethod = "dev"
	MethodUnknown  InstallMethod = "unknown"
)

func Detect(exePath string, installMethod string) InstallMethod {
	if installMethod != "" && installMethod != "direct" {
		return InstallMethod(installMethod)
	}
	if installMethod == "dev" {
		return MethodDev
	}

	path := strings.ToLower(filepath.Clean(exePath))
	switch {
	case strings.Contains(path, string(os.PathSeparator)+"cellar"+string(os.PathSeparator)+"muxdev"):
		return MethodHomebrew
	case strings.Contains(path, "homebrew"):
		return MethodHomebrew
	case strings.Contains(path, "scoop"+string(os.PathSeparator)+"apps"+string(os.PathSeparator)+"muxdev"):
		return MethodScoop
	case strings.Contains(path, "windowsapps"):
		return MethodWinget
	case strings.Contains(path, "winget"):
		return MethodWinget
	default:
		if method := detectGoInstall(path); method != "" {
			return method
		}
	}
	return MethodDirect
}

func detectGoInstall(path string) InstallMethod {
	for _, dir := range goBinDirs() {
		if dir == "" {
			continue
		}
		lower := strings.ToLower(filepath.Clean(dir))
		if strings.HasPrefix(path, lower) {
			return MethodGoInstall
		}
	}
	return ""
}

func goBinDirs() []string {
	dirs := []string{os.Getenv("GOBIN")}
	if gopath := os.Getenv("GOPATH"); gopath != "" {
		dirs = append(dirs, filepath.Join(gopath, "bin"))
	}
	if home, err := os.UserHomeDir(); err == nil {
		dirs = append(dirs, filepath.Join(home, "go", "bin"))
	}
	return dirs
}

func (m InstallMethod) UpgradeHint(latest string) string {
	module := fmt.Sprintf("github.com/%s/%s/cmd/muxdev", RepoOwner, RepoName)
	switch m {
	case MethodHomebrew:
		return "brew upgrade muxdev"
	case MethodScoop:
		return "scoop update muxdev"
	case MethodWinget:
		return "winget upgrade yarkingulacti.muxdev"
	case MethodGoInstall:
		return fmt.Sprintf("go install %s@v%s", module, strings.TrimPrefix(latest, "v"))
	case MethodDev:
		return fmt.Sprintf("go install %s@latest", module)
	default:
		return "muxdev update"
	}
}

func (m InstallMethod) SupportsSelfUpdate() bool {
	return m == MethodDirect || m == MethodUnknown
}

func CurrentExecutable() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(exe)
}

func PlatformAssetName(version, goos, goarch string) string {
	v := strings.TrimPrefix(version, "v")
	if goos == "windows" {
		return fmt.Sprintf("muxdev_%s_%s_%s.zip", v, goos, goarch)
	}
	return fmt.Sprintf("muxdev_%s_%s_%s.tar.gz", v, goos, goarch)
}

func CurrentPlatform() (goos, goarch string) {
	return runtime.GOOS, runtime.GOARCH
}
