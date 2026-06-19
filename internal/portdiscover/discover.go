package portdiscover

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const defaultTimeout = 12 * time.Second

var patterns = []struct {
	re       *regexp.Regexp
	priority int
}{
	{regexp.MustCompile(`(?i)local:\s+https?://[^:/\s]+:(\d{2,5})`), 100},
	{regexp.MustCompile(`(?i)(?:listening|ready|running|serving|started)(?:\s+\w+){0,4}\s+(?:on|at)\s+(?:https?://)?[^:\s]*:(\d{2,5})`), 90},
	{regexp.MustCompile(`(?i)(?:https?://)(?:localhost|127\.0\.0\.1|\[::1\]|0\.0\.0\.0):(\d{2,5})`), 80},
	{regexp.MustCompile(`(?i)(?:localhost|127\.0\.0\.1|\[::1\]|0\.0\.0\.0):(\d{2,5})`), 70},
	{regexp.MustCompile(`(?i)(?:port|addr(?:ess)?)\s*[:\s=]+['"]?(\d{2,5})`), 60},
}

// ParseLine extracts the best port match from a single log line.
func ParseLine(line string) string {
	type hit struct {
		port     string
		priority int
	}
	var best hit
	for _, p := range patterns {
		if m := p.re.FindStringSubmatch(line); len(m) > 1 && validPort(m[1]) {
			if p.priority > best.priority {
				best = hit{port: m[1], priority: p.priority}
			}
		}
	}
	return best.port
}

// Discover runs command briefly and parses stdout/stderr for a likely HTTP port.
func Discover(ctx context.Context, workDir, command string) (string, error) {
	command = strings.TrimSpace(command)
	if command == "" {
		return "", fmt.Errorf("empty command")
	}
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), defaultTimeout)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	if workDir != "" {
		cmd.Dir = workDir
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", err
	}

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("start command: %w", err)
	}

	type hit struct {
		port     string
		priority int
		order    int
	}

	var hits []hit
	order := 0
	scan := func(r io.Reader) {
		sc := bufio.NewScanner(r)
		for sc.Scan() {
			order++
			line := sc.Text()
			for _, p := range patterns {
				if m := p.re.FindStringSubmatch(line); len(m) > 1 && validPort(m[1]) {
					hits = append(hits, hit{port: m[1], priority: p.priority, order: order})
				}
			}
		}
		_ = sc.Err()
	}

	done := make(chan struct{}, 2)
	go func() {
		scan(stdout)
		done <- struct{}{}
	}()
	go func() {
		scan(stderr)
		done <- struct{}{}
	}()

	<-done
	<-done
	_ = cmd.Wait()

	if len(hits) == 0 {
		return "", nil
	}

	sort.Slice(hits, func(i, j int) bool {
		if hits[i].priority != hits[j].priority {
			return hits[i].priority > hits[j].priority
		}
		return hits[i].order > hits[j].order
	})

	return hits[0].port, nil
}

func validPort(raw string) bool {
	n, err := strconv.Atoi(raw)
	if err != nil {
		return false
	}
	return n >= 1024 && n <= 65535
}
