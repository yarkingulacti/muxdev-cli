package version

import (
	"encoding/json"
	"fmt"
)

var (
	Version       = "dev"
	Commit        = "none"
	Date          = "unknown"
	InstallMethod = "direct"
)

type Info struct {
	Version       string `json:"version"`
	Commit        string `json:"commit"`
	Date          string `json:"date"`
	InstallMethod string `json:"install_method"`
}

func InfoStruct() Info {
	return Info{
		Version:       Version,
		Commit:        Commit,
		Date:          Date,
		InstallMethod: InstallMethod,
	}
}

func String() string {
	if Version == "dev" {
		return "dev (local build)"
	}
	if Commit != "none" && Commit != "" {
		return fmt.Sprintf("%s (commit %s, %s)", Version, shortCommit(Commit), Date)
	}
	return Version
}

func Short() string {
	return Version
}

func JSON() (string, error) {
	data, err := json.MarshalIndent(InfoStruct(), "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func IsDev() bool {
	return Version == "dev"
}

func shortCommit(commit string) string {
	if len(commit) <= 7 {
		return commit
	}
	return commit[:7]
}
