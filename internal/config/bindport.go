package config

import (
	"regexp"
	"strconv"
	"strings"
)

var portTemplateKeyRE = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)(?::-[^}]*)?\}|\$([A-Za-z_][A-Za-z0-9_]*)`)

var commonBindPortEnvKeys = []string{
	"PORT",
	"BACKEND_PORT",
	"XTTS_PORT",
	"UI_PORT",
	"XTTS_UI_PORT",
	"OLLAMA_PORT",
}

// BindPortForService returns the TCP port a service is expected to bind, using the
// same env layers as runtime plus common shell fallbacks (e.g. XTTS_PORT when
// muxdev.yaml references ${BACKEND_PORT} but .env only defines XTTS_PORT).
func BindPortForService(workDir string, svc Service) int {
	if p := parsePortNumber(ResolveServicePort(nil, workDir, svc).Port); p > 0 {
		return p
	}

	env := ServiceRunEnv(workDir, svc)
	seen := make(map[string]bool)
	for _, key := range bindPortEnvKeys(svc.Port) {
		if seen[key] {
			continue
		}
		seen[key] = true
		if p := parsePortNumber(env[key]); p > 0 {
			return p
		}
	}
	for _, key := range commonBindPortEnvKeys {
		if seen[key] {
			continue
		}
		if p := parsePortNumber(env[key]); p > 0 {
			return p
		}
	}
	return 0
}

func bindPortEnvKeys(portTpl string) []string {
	keys := make([]string, 0, 4)
	for _, match := range portTemplateKeyRE.FindAllStringSubmatch(portTpl, -1) {
		key := match[1]
		if key == "" {
			key = match[2]
		}
		if key != "" {
			keys = append(keys, key)
		}
	}
	return keys
}

func parsePortNumber(raw string) int {
	raw = strings.TrimSpace(raw)
	if raw == "" || strings.Contains(raw, "$") || strings.Contains(raw, "{") {
		return 0
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 || n > 65535 {
		return 0
	}
	return n
}
