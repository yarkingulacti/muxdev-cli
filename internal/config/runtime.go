package config

import (
	"fmt"
	"strings"
)

type Runtime string

const (
	RuntimeSync  Runtime = "sync"
	RuntimeAsync Runtime = "async"
)

const DefaultRuntime = RuntimeSync

func ParseRuntime(raw string) (Runtime, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", string(RuntimeSync), "sequential", "synchronous":
		return RuntimeSync, nil
	case string(RuntimeAsync), "parallel", "asynchronous":
		return RuntimeAsync, nil
	default:
		return "", fmt.Errorf("unknown runtime %q (use sync or async)", raw)
	}
}

func (c *Config) ResolvedRuntime(override string) (Runtime, error) {
	if strings.TrimSpace(override) != "" {
		return ParseRuntime(override)
	}
	if strings.TrimSpace(c.Runtime) != "" {
		return ParseRuntime(c.Runtime)
	}
	return DefaultRuntime, nil
}
