package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

func Save(path string, cfg *Config) error {
	if err := cfg.Validate(); err != nil {
		return err
	}

	data, err := Format(cfg)
	if err != nil {
		return err
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

func Format(cfg *Config) ([]byte, error) {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("marshal config: %w", err)
	}
	return data, nil
}

func NormalizeServiceID(raw string) (string, error) {
	id := strings.TrimSpace(strings.ToLower(raw))
	id = strings.ReplaceAll(id, " ", "_")
	id = strings.ReplaceAll(id, "-", "_")
	if id == "" {
		return "", fmt.Errorf("service id is required")
	}
	for _, r := range id {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			continue
		}
		return "", fmt.Errorf("service id %q: use lowercase letters, numbers, underscores", raw)
	}
	if id[0] >= '0' && id[0] <= '9' {
		return "", fmt.Errorf("service id must start with a letter")
	}
	return id, nil
}

func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
