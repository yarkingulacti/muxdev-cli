package config

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

const DefaultFilename = "muxdev.yaml"

type Config struct {
	Name     string             `yaml:"name"`
	Subtitle string             `yaml:"subtitle"`
	Runtime  string             `yaml:"runtime,omitempty"`
	Services map[string]Service `yaml:"services"`
}

type Service struct {
	Label     string            `yaml:"label"`
	Command   string            `yaml:"command"`
	Port      string            `yaml:"port"`
	DependsOn []string          `yaml:"depends_on"`
	Env       map[string]string `yaml:"env"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func FindDefault(startDir string) (string, error) {
	dir := startDir
	for {
		candidate := strings.Join([]string{dir, DefaultFilename}, string(os.PathSeparator))
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
		parent := dirParent(dir)
		if parent == dir {
			return "", fmt.Errorf("%s not found from %s", DefaultFilename, startDir)
		}
		dir = parent
	}
}

func (c *Config) Validate() error {
	if strings.TrimSpace(c.Name) == "" {
		return fmt.Errorf("config: name is required")
	}
	if len(c.Services) == 0 {
		return fmt.Errorf("config: at least one service is required")
	}

	for id, svc := range c.Services {
		if strings.TrimSpace(svc.Command) == "" {
			return fmt.Errorf("config: service %q: command is required", id)
		}
		if strings.TrimSpace(svc.Label) == "" {
			return fmt.Errorf("config: service %q: label is required", id)
		}
		for _, dep := range svc.DependsOn {
			if dep == id {
				return fmt.Errorf("config: service %q: cannot depend on itself", id)
			}
			if _, ok := c.Services[dep]; !ok {
				return fmt.Errorf("config: service %q: unknown dependency %q", id, dep)
			}
		}
	}

	if _, err := c.SortedServiceIDs(); err != nil {
		return err
	}

	if _, err := c.ResolvedRuntime(""); err != nil {
		return fmt.Errorf("config: %w", err)
	}

	return nil
}

func (c *Config) SortedServiceIDs() ([]string, error) {
	return topologicalSort(c.Services)
}

func (c *Config) ResolveServices(focus []string) ([]string, error) {
	if len(focus) == 0 {
		return c.SortedServiceIDs()
	}

	ids := make([]string, 0, len(focus))
	seen := make(map[string]struct{}, len(focus))
	for _, raw := range focus {
		id := strings.TrimSpace(raw)
		if id == "" {
			continue
		}
		if _, ok := c.Services[id]; !ok {
			return nil, fmt.Errorf("unknown service %q", id)
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}

	if len(ids) == 0 {
		return nil, fmt.Errorf("no services selected")
	}

	closure := make(map[string]struct{})
	var visit func(string) error
	visit = func(id string) error {
		if _, ok := closure[id]; ok {
			return nil
		}
		svc, ok := c.Services[id]
		if !ok {
			return fmt.Errorf("unknown service %q", id)
		}
		for _, dep := range svc.DependsOn {
			if err := visit(dep); err != nil {
				return err
			}
		}
		closure[id] = struct{}{}
		return nil
	}

	for _, id := range ids {
		if err := visit(id); err != nil {
			return nil, err
		}
	}

	subset := make(map[string]Service, len(closure))
	for id := range closure {
		subset[id] = c.Services[id]
	}

	return topologicalSort(subset)
}

// OrderForStart returns service IDs with dependencies first.
// When the graph cannot be topologically sorted, falls back to stable ID order.
func OrderForStart(ids []string, services map[string]Service) []string {
	if len(ids) == 0 {
		return nil
	}
	subset := make(map[string]Service, len(ids))
	for _, id := range ids {
		if svc, ok := services[id]; ok {
			subset[id] = svc
		}
	}
	sorted, err := topologicalSort(subset)
	if err != nil {
		return stablePriorityOrder(ids)
	}
	return sorted
}

func stablePriorityOrder(ids []string) []string {
	out := append([]string(nil), ids...)
	slices.Sort(out)
	return out
}

func topologicalSort(services map[string]Service) ([]string, error) {
	inDegree := make(map[string]int, len(services))
	dependents := make(map[string][]string, len(services))

	for id := range services {
		inDegree[id] = 0
	}
	for id, svc := range services {
		for _, dep := range svc.DependsOn {
			if _, ok := services[dep]; !ok {
				continue
			}
			inDegree[id]++
			dependents[dep] = append(dependents[dep], id)
		}
	}

	queue := make([]string, 0, len(services))
	for id, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, id)
		}
	}
	slices.Sort(queue)

	sorted := make([]string, 0, len(services))
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		sorted = append(sorted, id)

		for _, child := range dependents[id] {
			inDegree[child]--
			if inDegree[child] == 0 {
				queue = append(queue, child)
			}
		}
		slices.Sort(queue)
	}

	if len(sorted) != len(services) {
		return nil, fmt.Errorf("config: cyclic service dependencies detected")
	}

	return sorted, nil
}

func dirParent(path string) string {
	path = strings.TrimRight(path, string(os.PathSeparator))
	if path == "" {
		return string(os.PathSeparator)
	}
	idx := strings.LastIndex(path, string(os.PathSeparator))
	if idx <= 0 {
		return string(os.PathSeparator)
	}
	return path[:idx]
}
