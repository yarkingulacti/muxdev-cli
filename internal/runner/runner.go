package runner

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/yarkingulacti/muxdev-cli/internal/config"
	"github.com/yarkingulacti/muxdev-cli/internal/platform"
)

type LineHandler func(label string, stderr bool, text string)

type Context struct {
	WorkDir    string
	Stdout     io.Writer
	Stderr     io.Writer
	OnLine     LineHandler
	CancelFunc context.CancelFunc
}

type Runner struct {
	cfg        *config.Config
	serviceIDs []string
}

func New(cfg *config.Config, serviceIDs []string) *Runner {
	return &Runner{cfg: cfg, serviceIDs: serviceIDs}
}

func (r *Runner) Run(ctx Context) error {
	rootCtx, cancel := context.WithCancel(context.Background())
	if ctx.CancelFunc != nil {
		original := cancel
		cancel = func() {
			original()
			ctx.CancelFunc()
		}
	}
	defer cancel()

	if ctx.OnLine == nil {
		go platform.HandleInterrupt(cancel)
	}

	var wg sync.WaitGroup
	errCh := make(chan error, len(r.serviceIDs))

	for _, id := range r.serviceIDs {
		wg.Add(1)
		go func(serviceID string) {
			defer wg.Done()
			if err := r.runService(rootCtx, ctx, serviceID); err != nil {
				errCh <- err
				cancel()
			}
		}(id)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-rootCtx.Done():
		<-done
	case <-done:
	}

	close(errCh)
	for err := range errCh {
		if err != nil {
			return err
		}
	}

	if rootCtx.Err() == context.Canceled {
		return nil
	}
	return rootCtx.Err()
}

func (r *Runner) runService(ctx context.Context, runCtx Context, serviceID string) error {
	svc := r.cfg.Services[serviceID]
	cmd := exec.CommandContext(ctx, platform.ShellCommand(), platform.ShellArgs(svc.Command)...)
	cmd.Dir = runCtx.WorkDir
	cmd.Env = mergeEnv(os.Environ(), svc.Env)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("service %q: stdout pipe: %w", serviceID, err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("service %q: stderr pipe: %w", serviceID, err)
	}

	platform.ConfigureCommand(cmd)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("service %q: start: %w", serviceID, err)
	}

	label := serviceLabel(serviceID, svc.Label)
	var streamWG sync.WaitGroup
	streamWG.Add(2)
	go func() {
		defer streamWG.Done()
		streamLines(ctx, runCtx, label, false, stdout)
	}()
	go func() {
		defer streamWG.Done()
		streamLines(ctx, runCtx, label+"!", true, stderr)
	}()

	waitErr := cmd.Wait()
	streamWG.Wait()

	if ctx.Err() != nil {
		return nil
	}
	if waitErr != nil {
		return fmt.Errorf("service %q: %w", serviceID, waitErr)
	}
	return nil
}

func streamLines(ctx context.Context, runCtx Context, label string, stderr bool, r io.Reader) {
	scanner := bufio.NewScanner(r)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
		}
		line := scanner.Text()
		emitLine(runCtx, label, stderr, line)
	}
}

func emitLine(ctx Context, label string, stderr bool, line string) {
	if ctx.OnLine != nil {
		ctx.OnLine(label, stderr, line)
		return
	}
	out := ctx.Stdout
	prefix := label
	if stderr {
		out = ctx.Stderr
	}
	if out == nil {
		return
	}
	fmt.Fprintf(out, "[%s] %s\n", prefix, line)
}

func serviceLabel(id, label string) string {
	if strings.TrimSpace(label) == "" {
		return id
	}
	return label
}

func mergeEnv(base []string, extra map[string]string) []string {
	if len(extra) == 0 {
		return base
	}
	out := append([]string(nil), base...)
	for key, value := range extra {
		out = append(out, key+"="+value)
	}
	return out
}
