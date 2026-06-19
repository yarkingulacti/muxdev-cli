package runner

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"time"

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
	runtime    config.Runtime
}

func New(cfg *config.Config, serviceIDs []string, runtime config.Runtime) *Runner {
	if runtime == "" {
		runtime = config.DefaultRuntime
	}
	return &Runner{cfg: cfg, serviceIDs: serviceIDs, runtime: runtime}
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

	order := config.OrderForStart(r.serviceIDs, r.cfg.Services)
	if r.runtime == config.RuntimeAsync {
		return r.runAsync(rootCtx, ctx, order, cancel)
	}
	return r.runSync(rootCtx, ctx, order, cancel)
}

func (r *Runner) runSync(rootCtx context.Context, ctx Context, order []string, cancel context.CancelFunc) error {
	handles := make([]*serviceHandle, 0, len(order))
	defer r.stopAll(handles)

	for _, id := range order {
		if rootCtx.Err() != nil {
			break
		}
		handle, err := r.launchService(rootCtx, ctx, id)
		if err != nil {
			cancel()
			return err
		}
		handles = append(handles, handle)
	}

	return r.waitHandles(rootCtx, cancel, handles)
}

func (r *Runner) runAsync(rootCtx context.Context, ctx Context, order []string, cancel context.CancelFunc) error {
	handles := make([]*serviceHandle, 0, len(order))
	var mu sync.Mutex
	var launchErr error

	var wg sync.WaitGroup
	for _, id := range order {
		wg.Add(1)
		go func(serviceID string) {
			defer wg.Done()
			handle, err := r.launchService(rootCtx, ctx, serviceID)
			if err != nil {
				mu.Lock()
				if launchErr == nil {
					launchErr = err
					cancel()
				}
				mu.Unlock()
				return
			}
			mu.Lock()
			handles = append(handles, handle)
			mu.Unlock()
		}(id)
	}
	wg.Wait()

	defer r.stopAll(handles)

	if launchErr != nil {
		return launchErr
	}

	mu.Lock()
	h := append([]*serviceHandle(nil), handles...)
	mu.Unlock()
	return r.waitHandles(rootCtx, cancel, h)
}

func (r *Runner) stopAll(handles []*serviceHandle) {
	for _, handle := range handles {
		handle.stop()
	}
}

func (r *Runner) waitHandles(rootCtx context.Context, cancel context.CancelFunc, handles []*serviceHandle) error {
	if len(handles) == 0 {
		return rootCtx.Err()
	}

	var wg sync.WaitGroup
	errCh := make(chan error, len(handles))

	for _, handle := range handles {
		wg.Add(1)
		go func(h *serviceHandle) {
			defer wg.Done()
			if err := h.wait(rootCtx); err != nil {
				errCh <- err
				cancel()
			}
		}(handle)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-rootCtx.Done():
		for _, handle := range handles {
			handle.stop()
		}
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

type serviceHandle struct {
	id       string
	cmd      *exec.Cmd
	streamWG sync.WaitGroup
	stopped  bool
}

func (r *Runner) launchService(ctx context.Context, runCtx Context, serviceID string) (*serviceHandle, error) {
	svc := r.cfg.Services[serviceID]
	cmd := exec.CommandContext(ctx, platform.ShellCommand(), platform.ShellArgs(svc.Command)...)
	cmd.Dir = runCtx.WorkDir
	cmd.Env = envMapToSlice(config.ServiceRunEnv(runCtx.WorkDir, svc))

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("service %q: stdout pipe: %w", serviceID, err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("service %q: stderr pipe: %w", serviceID, err)
	}

	platform.ConfigureCommand(cmd)

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("service %q: start: %w", serviceID, err)
	}

	label := serviceLabel(serviceID, svc.Label)
	handle := &serviceHandle{id: serviceID, cmd: cmd}
	handle.streamWG.Add(2)
	go func() {
		defer handle.streamWG.Done()
		streamLines(ctx, runCtx, label, false, stdout)
	}()
	go func() {
		defer handle.streamWG.Done()
		streamLines(ctx, runCtx, label+"!", true, stderr)
	}()

	return handle, nil
}

func (h *serviceHandle) stop() {
	if h.stopped {
		return
	}
	h.stopped = true
	platform.StopProcessGroup(h.cmd)
}

func (h *serviceHandle) wait(ctx context.Context) error {
	if h.cmd == nil {
		return nil
	}

	waitDone := make(chan error, 1)
	go func() {
		waitDone <- h.cmd.Wait()
	}()

	select {
	case <-ctx.Done():
		h.stop()
		select {
		case <-waitDone:
		case <-time.After(5 * time.Second):
			platform.KillProcessGroup(h.cmd)
			<-waitDone
		}
		h.streamWG.Wait()
		return nil
	case err := <-waitDone:
		h.streamWG.Wait()
		if ctx.Err() != nil {
			return nil
		}
		if err != nil {
			return fmt.Errorf("service %q: %w", h.id, err)
		}
		return nil
	}
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

func envMapToSlice(env map[string]string) []string {
	out := make([]string, 0, len(env))
	for key, value := range env {
		out = append(out, key+"="+value)
	}
	return out
}
