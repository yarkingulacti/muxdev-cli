package update

import (
	"fmt"
	"strings"
)

func httpStatusError(status int, body string) error {
	summary := summarizeHTTPBody(body)
	if summary == "" {
		return fmt.Errorf("http %d", status)
	}
	return fmt.Errorf("http %d: %s", status, summary)
}

func summarizeHTTPBody(body string) string {
	body = strings.TrimSpace(body)
	if body == "" {
		return ""
	}
	lower := strings.ToLower(body)
	if strings.HasPrefix(lower, "<!doctype") || strings.HasPrefix(lower, "<html") {
		return "unexpected HTML response"
	}
	if len(body) > 200 {
		return body[:200] + "..."
	}
	return body
}
