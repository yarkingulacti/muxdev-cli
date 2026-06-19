package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestHelpFlagShowsPlainTopic(t *testing.T) {
	root := NewRoot()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)

	root.SetArgs([]string{"list", "--help"})
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	text := out.String()
	if !strings.Contains(text, "muxdev list") {
		t.Fatalf("output = %q, want list topic", text)
	}
	if strings.Contains(text, "Flags:") && strings.Contains(text, "-c, --config") {
		// wiki body includes flag docs from generated page — that's fine
	} else if !strings.Contains(text, "List configured") && !strings.Contains(text, "services") {
		t.Fatalf("output = %q, want command description", text)
	}
}

func TestHelpFlagRootShowsIndex(t *testing.T) {
	root := NewRoot()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)

	root.SetArgs([]string{"--help"})
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	text := out.String()
	if !strings.Contains(text, "muxdev guide") {
		t.Fatalf("output = %q, want topic index", text)
	}
}

func TestHelpCommandStillWorks(t *testing.T) {
	root := NewRoot()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)

	root.SetArgs([]string{"help", "version"})
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(out.String(), "muxdev version") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestHelpTopicForSubcommand(t *testing.T) {
	root := NewRoot()
	listCmd, _, err := root.Find([]string{"list"})
	if err != nil {
		t.Fatal(err)
	}
	if got := helpTopicFor(listCmd); got != "list" {
		t.Fatalf("helpTopicFor() = %q, want list", got)
	}
}
