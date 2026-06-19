package config

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var shellDefaultPattern = regexp.MustCompile(`\$\{([^}:]+):-([^}]*)\}`)

// ListingEnv builds the environment used to expand ${VAR} placeholders in muxdev list.
// OS variables take precedence; then env_source files (bash source); then .env.example; then per-service env.
func ListingEnv(workDir string, envSource []string, svcEnv map[string]string) map[string]string {
	env := envFromOS()
	if workDir != "" {
		mergeSourcedEnv(env, workDir, envSource)
	}
	for key, value := range svcEnv {
		env[key] = value
	}
	return env
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

func defaultEnvSource(custom []string) []string {
	if len(custom) > 0 {
		return custom
	}
	return []string{".env", ".env.local"}
}

func mergeSourcedEnv(env map[string]string, workDir string, envSource []string) {
	sources := defaultEnvSource(envSource)
	fileEnv, err := sourceEnvFiles(workDir, sources)
	if err != nil || len(fileEnv) == 0 {
		mergeParsedDotenvFiles(env, workDir, sources)
	} else {
		for key, value := range fileEnv {
			if _, exists := env[key]; !exists {
				env[key] = value
			}
		}
	}

	for key, value := range loadEnvExampleFile(filepath.Join(workDir, ".env.example")) {
		if _, exists := env[key]; !exists {
			env[key] = value
		}
	}
}

func mergeParsedDotenvFiles(env map[string]string, workDir string, sources []string) {
	fileEnv := make(map[string]string)
	for _, name := range sources {
		mergeParsedEnv(fileEnv, workDir, name)
	}
	for key, value := range fileEnv {
		if _, exists := env[key]; !exists {
			env[key] = value
		}
	}
}

func sourceEnvFiles(workDir string, sources []string) (map[string]string, error) {
	if _, err := exec.LookPath("bash"); err != nil {
		return nil, err
	}

	var script bytes.Buffer
	script.WriteString("cd ")
	writeShellQuoted(&script, workDir)
	script.WriteString("\nset -a\n")
	found := false
	for _, source := range sources {
		path, err := resolveEnvSourcePath(workDir, source)
		if err != nil {
			continue
		}
		if _, err := os.Stat(path); err != nil {
			continue
		}
		found = true
		script.WriteString("source ")
		writeShellQuoted(&script, path)
		script.WriteString("\n")
	}
	script.WriteString("set +a\nenv -0\n")
	if !found {
		return nil, nil
	}

	cmd := exec.Command("bash", "-c", script.String())
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseNullSeparatedEnv(out), nil
}

func resolveEnvSourcePath(workDir, rel string) (string, error) {
	rel = strings.TrimSpace(rel)
	if rel == "" {
		return "", fmt.Errorf("empty env source path")
	}
	if filepath.IsAbs(rel) {
		return "", fmt.Errorf("absolute env_source paths are not allowed: %s", rel)
	}
	clean := filepath.Clean(rel)
	if clean == ".." || strings.HasPrefix(clean, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("env_source path escapes project root: %s", rel)
	}
	full := filepath.Join(workDir, clean)
	relToRoot, err := filepath.Rel(workDir, full)
	if err != nil || strings.HasPrefix(relToRoot, "..") {
		return "", fmt.Errorf("env_source path escapes project root: %s", rel)
	}
	return full, nil
}

func writeShellQuoted(b *bytes.Buffer, value string) {
	b.WriteByte('\'')
	b.WriteString(strings.ReplaceAll(value, "'", `'"\''"`))
	b.WriteByte('\'')
}

func parseNullSeparatedEnv(raw []byte) map[string]string {
	out := make(map[string]string)
	for part := range bytes.SplitSeq(raw, []byte{0}) {
		if len(part) == 0 {
			continue
		}
		key, value, ok := bytes.Cut(part, []byte{'='})
		if !ok || len(key) == 0 {
			continue
		}
		out[string(key)] = string(value)
	}
	return out
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

func mergeParsedEnv(dst map[string]string, workDir, name string) {
	parsed, err := parseEnvFile(filepath.Join(workDir, name))
	if err != nil || len(parsed) == 0 {
		return
	}
	for key, value := range parsed {
		dst[key] = value
	}
}

func loadEnvFile(path string) map[string]string {
	parsed, err := parseEnvFile(path)
	if err != nil {
		return nil
	}
	return parsed
}

func loadEnvExampleFile(path string) map[string]string {
	out := loadEnvFile(path)
	if out == nil {
		out = make(map[string]string)
	}
	commented, err := parseCommentedEnvDefaults(path)
	if err != nil {
		return out
	}
	for key, value := range commented {
		if _, exists := out[key]; !exists {
			out[key] = value
		}
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
