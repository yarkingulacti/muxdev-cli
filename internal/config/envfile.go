package config

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var shellDefaultPattern = regexp.MustCompile(`\$\{([^}:]+):-([^}]*)\}`)
var portVarPattern = regexp.MustCompile(`\$\{([^}:]+)(?::-([^}]*))?\}`)
var bareVarPattern = regexp.MustCompile(`\$([A-Za-z_][A-Za-z0-9_]*)`)

// ServicePortResolution holds an expanded port and where it was resolved from.
type ServicePortResolution struct {
	Port   string
	Source string
}

// PortListingEnv builds the environment used to expand port placeholders.
// Priority (highest first): shell → .env.local → .env → .env.example → commented .env.example → service env.
func PortListingEnv(workDir string, svcEnv map[string]string) map[string]string {
	env, _ := PortListingEnvWithSources(workDir, svcEnv)
	return env
}

// PortListingEnvWithSources is like PortListingEnv but also returns the winning source per variable.
func PortListingEnvWithSources(workDir string, svcEnv map[string]string) (map[string]string, map[string]string) {
	env := make(map[string]string)
	sources := make(map[string]string)
	for _, layer := range portEnvLayers(workDir, svcEnv) {
		for key, value := range layer.data {
			env[key] = value
			sources[key] = layer.name
		}
	}
	expandEnvReferences(env)
	return env, sources
}

// ListingEnv is an alias for PortListingEnv.
func ListingEnv(workDir string, svcEnv map[string]string) map[string]string {
	return PortListingEnv(workDir, svcEnv)
}

// ServiceRunEnv returns env vars for starting a service process.
// Resolution order matches port listing: service env < .env files < current shell.
func ServiceRunEnv(workDir string, svc Service) map[string]string {
	return PortListingEnv(workDir, svc.Env)
}

// ExpandServicePort resolves a service port field using layered env sources.
func ExpandServicePort(cfg *Config, workDir string, svc Service) string {
	return ResolveServicePort(cfg, workDir, svc).Port
}

// ResolveServicePort expands a service port and reports where the value came from.
func ResolveServicePort(_ *Config, workDir string, svc Service) ServicePortResolution {
	portTpl := strings.TrimSpace(svc.Port)
	if portTpl == "" {
		return ServicePortResolution{}
	}
	if !containsEnvRef(portTpl) {
		return ServicePortResolution{Port: portTpl, Source: "muxdev.yaml"}
	}

	env, sources := PortListingEnvWithSources(workDir, svc.Env)
	return ServicePortResolution{
		Port:   ExpandEnv(portTpl, env),
		Source: resolvePortSource(portTpl, env, sources),
	}
}

// ExpandEnv replaces ${VAR}, ${VAR:-default}, and $VAR using env; unknown keys stay as ${KEY}.
func ExpandEnv(value string, env map[string]string) string {
	if strings.TrimSpace(value) == "" {
		return value
	}
	value = expandShellDefaults(value, env)
	return os.Expand(value, func(key string) string {
		if v, ok := env[key]; ok {
			return v
		}
		return "${" + key + "}"
	})
}

func buildLayeredEnv(layers ...map[string]string) map[string]string {
	env := make(map[string]string)
	for _, layer := range layers {
		for key, value := range layer {
			env[key] = value
		}
	}
	return env
}

type envLayer struct {
	name string
	data map[string]string
}

func portEnvLayers(workDir string, svcEnv map[string]string) []envLayer {
	layers := []envLayer{{name: "env", data: svcEnv}}
	if workDir != "" {
		examplePath := filepath.Join(workDir, ".env.example")
		layers = append(layers,
			envLayer{name: ".env.example#", data: envExampleCommented(examplePath)},
			envLayer{name: ".env.example", data: envExampleExplicit(examplePath)},
			envLayer{name: ".env", data: parsedEnvFile(workDir, ".env")},
			envLayer{name: ".env.local", data: parsedEnvFile(workDir, ".env.local")},
		)
	}
	layers = append(layers, envLayer{name: "shell", data: envFromOS()})
	return layers
}

func containsEnvRef(value string) bool {
	return portVarPattern.MatchString(value) || strings.Contains(value, "$")
}

func resolvePortSource(portTpl string, env map[string]string, sources map[string]string) string {
	for _, match := range portVarPattern.FindAllStringSubmatch(portTpl, -1) {
		key := match[1]
		fallback := match[2]
		hasDefault := strings.Contains(portTpl, "${"+key+":-")

		if v, ok := env[key]; ok && v != "" {
			if src, ok := sources[key]; ok {
				return src
			}
		}
		if hasDefault && fallback != "" {
			return "default"
		}
	}

	for _, match := range bareVarPattern.FindAllStringSubmatch(portTpl, -1) {
		key := match[1]
		if v, ok := env[key]; ok && v != "" {
			if src, ok := sources[key]; ok {
				return src
			}
		}
	}

	return "?"
}

func expandEnvReferences(env map[string]string) {
	for range 8 {
		changed := false
		for key, value := range env {
			expanded := ExpandEnv(value, env)
			if expanded != value {
				env[key] = expanded
				changed = true
			}
		}
		if !changed {
			break
		}
	}
}

func envExampleExplicit(path string) map[string]string {
	parsed, err := parseEnvFile(path)
	if err != nil || len(parsed) == 0 {
		return map[string]string{}
	}
	return parsed
}

func envExampleCommented(path string) map[string]string {
	parsed, err := parseCommentedEnvDefaults(path)
	if err != nil || len(parsed) == 0 {
		return map[string]string{}
	}
	return parsed
}

func parsedEnvFile(workDir, name string) map[string]string {
	parsed, err := parseEnvFile(filepath.Join(workDir, name))
	if err != nil || len(parsed) == 0 {
		return map[string]string{}
	}
	return parsed
}

func expandShellDefaults(value string, env map[string]string) string {
	return shellDefaultPattern.ReplaceAllStringFunc(value, func(match string) string {
		sub := shellDefaultPattern.FindStringSubmatch(match)
		if len(sub) != 3 {
			return match
		}
		key, fallback := sub[1], sub[2]
		if v, ok := env[key]; ok && v != "" {
			return v
		}
		return fallback
	})
}

func envFromOS() map[string]string {
	out := make(map[string]string)
	for _, line := range os.Environ() {
		key, value, ok := strings.Cut(line, "=")
		if !ok || key == "" {
			continue
		}
		out[key] = value
	}
	return out
}

func parseCommentedEnvDefaults(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	out := make(map[string]string)
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if !strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimSpace(strings.TrimPrefix(line, "#"))
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		out[key] = unquoteEnvValue(strings.TrimSpace(value))
	}
	return out, sc.Err()
}

func parseEnvFile(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	out := make(map[string]string)
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		out[key] = unquoteEnvValue(strings.TrimSpace(value))
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func unquoteEnvValue(value string) string {
	if len(value) >= 2 {
		if (value[0] == '"' && value[len(value)-1] == '"') ||
			(value[0] == '\'' && value[len(value)-1] == '\'') {
			return value[1 : len(value)-1]
		}
	}
	return value
}
