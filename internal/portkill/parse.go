package portkill

import (
	"fmt"
	"regexp"
	"strconv"
)

var (
	eaddrInUseRE       = regexp.MustCompile(`(?i)EADDRINUSE(?:.*?)(?:::(\d{1,5})|[^0-9](\d{1,5})\s*$|port\s+(\d{1,5}))`)
	portAlreadyInUseRE = regexp.MustCompile(`(?i)port\s+(\d{1,5})\s+is\s+already\s+in\s+use`)
	portInUseRE        = regexp.MustCompile(`(?i)port\s+(\d{1,5})\s+is in use`)
)

// Conflict describes a detected port collision from service logs.
type Conflict struct {
	Port  int
	Fatal bool // true when the service crashed (EADDRINUSE)
}

// ParseConflict extracts port conflict info from a log line.
func ParseConflict(line string) (Conflict, bool) {
	if m := eaddrInUseRE.FindStringSubmatch(line); len(m) > 0 {
		if port := firstPort(m[1:]); port > 0 {
			return Conflict{Port: port, Fatal: true}, true
		}
	}
	if m := portAlreadyInUseRE.FindStringSubmatch(line); len(m) > 1 {
		if port, err := strconv.Atoi(m[1]); err == nil && validPort(port) {
			return Conflict{Port: port, Fatal: true}, true
		}
	}
	if m := portInUseRE.FindStringSubmatch(line); len(m) > 1 {
		if port, err := strconv.Atoi(m[1]); err == nil && validPort(port) {
			return Conflict{Port: port, Fatal: false}, true
		}
	}
	return Conflict{}, false
}

func firstPort(groups []string) int {
	for _, g := range groups {
		if g == "" {
			continue
		}
		n, err := strconv.Atoi(g)
		if err != nil || !validPort(n) {
			continue
		}
		return n
	}
	return 0
}

func validPort(n int) bool {
	return n >= 1 && n <= 65535
}

func (c Conflict) Message() string {
	if c.Fatal {
		return fmt.Sprintf("Port %d is already in use", c.Port)
	}
	return fmt.Sprintf("Port %d is in use (service may pick another port)", c.Port)
}
