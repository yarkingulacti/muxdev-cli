package tui

import (
	"strings"
	"testing"
)

func TestRenderInitProgressHighlightsCurrentStep(t *testing.T) {
	got := renderInitProgress(phaseCfgName)
	for _, want := range []string{"✓ Start", "› Project", "Services", "Review"} {
		if !strings.Contains(got, want) {
			t.Fatalf("renderInitProgress() missing %q:\n%s", want, got)
		}
	}
}

func TestRenderInitDetectionCalloutOmitsEmpty(t *testing.T) {
	if got := renderInitDetectionCallout("", "", ""); got != "" {
		t.Fatalf("expected empty callout, got %q", got)
	}
}

func TestRenderInitDetectionCalloutIncludesHints(t *testing.T) {
	got := renderInitDetectionCallout("My App", "npm run dev", "./muxdev.yaml")
	for _, want := range []string{"My App", "npm run dev", "muxdev.yaml"} {
		if !strings.Contains(got, want) {
			t.Fatalf("renderInitDetectionCallout() missing %q:\n%s", want, got)
		}
	}
}

func TestShowConfigureHeaderSkipsWelcome(t *testing.T) {
	m := configureModel{init: true, phase: phaseCfgWelcome}
	if m.showConfigureHeader() {
		t.Fatal("welcome should hide standard header")
	}
	m.phase = phaseCfgName
	if !m.showConfigureHeader() {
		t.Fatal("name step should show standard header")
	}
}

func TestInitStepForPhase(t *testing.T) {
	step, label := initStepForPhase(phaseCfgServiceCommand)
	if step != 3 || label != "Services" {
		t.Fatalf("initStepForPhase() = (%d, %q), want (3, Services)", step, label)
	}
}
