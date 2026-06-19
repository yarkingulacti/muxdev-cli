package tui

import (
	"testing"

	"github.com/yarkingulacti/muxdev-cli/internal/runner"
)

func TestCanWaitForRunnerShutdown(t *testing.T) {
	m := runnerModel{cancel: func() {}}
	if !m.canWaitForRunnerShutdown() {
		t.Fatal("active cancel should wait for shutdown")
	}
	m = runnerModel{done: true}
	if m.canWaitForRunnerShutdown() {
		t.Fatal("idle finished runner should not wait")
	}
}

func TestRequestShutdownGracefulWhenIdleQuits(t *testing.T) {
	doneCh := make(chan runDoneMsg, 1)
	m := runnerModel{
		done:   true,
		doneCh: doneCh,
		cancel: nil,
	}

	_, cmd := m.requestShutdown(false)
	if cmd == nil {
		t.Fatal("expected quit when runner is already idle")
	}
}

func TestRequestShutdownGracefulEscalatesWhenAlreadyShuttingDown(t *testing.T) {
	shutdown := &runner.ShutdownRequest{}
	m := runnerModel{
		shuttingDown: true,
		shutdown:     shutdown,
		cancel:       func() {},
	}

	_, cmd := m.requestShutdown(false)
	if cmd == nil {
		t.Fatal("second graceful quit should force exit")
	}
	if !shutdown.Forceful {
		t.Fatal("expected forceful escalation")
	}
}

func TestRequestShutdownGracefulWaitsForActiveRunner(t *testing.T) {
	shutdown := &runner.ShutdownRequest{}
	m := runnerModel{
		shutdown: shutdown,
		cancel:   func() {},
	}

	next, cmd := m.requestShutdown(false)
	if cmd != nil {
		t.Fatal("graceful shutdown should wait for runner completion")
	}
	if !next.shuttingDown {
		t.Fatal("expected shuttingDown")
	}
	if shutdown.Forceful {
		t.Fatal("graceful shutdown should not set forceful flag")
	}
}

func TestRequestShutdownForcefulQuitsImmediately(t *testing.T) {
	shutdown := &runner.ShutdownRequest{}
	m := runnerModel{
		shutdown: shutdown,
		cancel:   func() {},
	}

	_, cmd := m.requestShutdown(true)
	if cmd == nil {
		t.Fatal("expected immediate quit")
	}
	if !shutdown.Forceful {
		t.Fatal("expected forceful shutdown flag")
	}
}

func TestAttachDoneMsgQuitsWhenShuttingDown(t *testing.T) {
	m := runnerModel{
		attached:     true,
		shuttingDown: true,
		attachCancel: func() {},
	}

	next, cmd := m.Update(attachDoneMsg{})
	if cmd == nil {
		t.Fatal("expected quit after attach ends during shutdown")
	}
	if next.(runnerModel).attached {
		t.Fatal("expected attach cleared")
	}
}

func TestPortKillMsgDoesNotRestartWhenShuttingDown(t *testing.T) {
	m := runnerModel{shuttingDown: true}

	_, cmd := m.Update(portKillMsg{port: 3000})
	if cmd == nil {
		t.Fatal("expected quit during shutdown")
	}
}

func TestRunDoneMsgQuitsWhenShuttingDown(t *testing.T) {
	m := runnerModel{
		shuttingDown: true,
		cancel:       func() {},
	}

	_, cmd := m.Update(runDoneMsg{})
	if cmd == nil {
		t.Fatal("expected quit on run done while shutting down")
	}
}
