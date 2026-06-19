//go:build !windows

package portkill

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

// AttachProcess streams output from the given PID until ctx is cancelled or the process exits.
func AttachProcess(ctx context.Context, pid int, onLine LineHandler) error {
	if onLine == nil {
		return fmt.Errorf("line handler is required")
	}

	var wg sync.WaitGroup
	errCh := make(chan error, 2)

	stream := func(fd int, stderr bool) {
		defer wg.Done()
		if err := streamFD(ctx, pid, fd, stderr, onLine); err != nil {
			select {
			case errCh <- err:
			default:
			}
		}
	}

	wg.Add(2)
	go stream(1, false)
	go stream(2, true)

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		return nil
	case err := <-errCh:
		return err
	case <-done:
		return nil
	}
}

func streamFD(ctx context.Context, pid, fd int, stderr bool, onLine LineHandler) error {
	path := fmt.Sprintf("/proc/%d/fd/%d", pid, fd)
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	reader := bufio.NewReader(f)
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		line, err := readLine(ctx, reader)
		if err != nil {
			if err == io.EOF || ctx.Err() != nil {
				return nil
			}
			return err
		}
		if line == "" {
			continue
		}
		onLine(stderr, line)
	}
}

func readLine(ctx context.Context, r *bufio.Reader) (string, error) {
	type result struct {
		line string
		err  error
	}
	ch := make(chan result, 1)
	go func() {
		line, err := r.ReadString('\n')
		ch <- result{line: strings.TrimRight(line, "\r\n"), err: err}
	}()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case res := <-ch:
		return res.line, res.err
	}
}
