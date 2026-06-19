package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/term"

	"github.com/yarkingulacti/muxdev-cli/internal/config"
	"github.com/yarkingulacti/muxdev-cli/internal/logs"
	"github.com/yarkingulacti/muxdev-cli/internal/runner"
	"github.com/yarkingulacti/muxdev-cli/internal/tui"
	"github.com/yarkingulacti/muxdev-cli/internal/update"
)

func run(opts Options, focus string) error {
	cfgPath, err := resolveConfigPath(opts.ConfigPath)
	if err != nil {
		return err
	}

	if !config.Exists(cfgPath) {
		if opts.ConfigPath == "" {
			return fmt.Errorf("%s not found (run `muxdev init` or pass --config)", cfgPath)
		}
		interactive := !opts.NoInteractive && isTerminal(os.Stdout)
		if !interactive {
			return fmt.Errorf("%s not found (run `muxdev init` to create it)", cfgPath)
		}
		if err := runInitWizard(cfgPath); err != nil {
			return err
		}
		if !config.Exists(cfgPath) {
			return nil
		}
	}

	workDir, err := resolveWorkDir(cfgPath)
	if err != nil {
		return err
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return err
	}

	focusIDs := parseFocus(focus)
	interactive := !opts.NoInteractive && isTerminal(os.Stdout)

	runtime, err := cfg.ResolvedRuntime(opts.Runtime)
	if err != nil {
		return err
	}

	if interactive {
		hint := ""
		if update.ShouldCheckOnStart(24 * time.Hour) {
			hint = update.StartupHint()
			go backgroundUpdateCheck()
		}
		err := tui.Run(tui.Options{
			Cfg:        cfg,
			ConfigPath: cfgPath,
			Focus:      focusIDs,
			WorkDir:    workDir,
			UpdateHint: hint,
			Runtime:    runtime,
		})
		if errors.Is(err, tui.ErrAborted) {
			return nil
		}
		return err
	}

	serviceIDs, err := cfg.ResolveServices(focusIDs)
	if err != nil {
		return err
	}

	r := runner.New(cfg, serviceIDs, runtime)
	session, err := logs.StartSession(workDir, cfgPath, serviceIDs, string(runtime))
	if err != nil {
		session = nil
	}

	runErr := r.Run(runner.Context{
		WorkDir: workDir,
		OnLine: func(label string, stderr bool, text string) {
			if session != nil {
				_ = session.Append(label, text)
			}
			out := os.Stdout
			if stderr {
				out = os.Stderr
			}
			fmt.Fprintf(out, "[%s] %s\n", label, text)
		},
	})
	if session != nil {
		_ = session.Finish(runErr)
	}
	return runErr
}

func backgroundUpdateCheck() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	result, err := update.Check(ctx, update.CheckOptions{Channel: update.ChannelStable})
	if err != nil {
		return
	}
	_ = update.WriteCache(update.CacheEntry{
		CheckedAt:       time.Now(),
		Current:         result.Current,
		Latest:          result.Latest,
		UpdateAvailable: result.UpdateAvailable,
	})
}

func isTerminal(out *os.File) bool {
	return term.IsTerminal(int(out.Fd()))
}

func resolveConfigPath(explicit string) (string, error) {
	if explicit != "" {
		return explicit, nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}
	if path, err := config.FindDefault(cwd); err == nil {
		return path, nil
	}
	return "", fmt.Errorf("%s not found from %s", config.DefaultFilename, cwd)
}

func resolveWorkDir(cfgPath string) (string, error) {
	abs, err := filepath.Abs(cfgPath)
	if err != nil {
		return "", fmt.Errorf("resolve work directory: %w", err)
	}
	return filepath.Dir(abs), nil
}

func runInitWizard(path string) error {
	workDir, err := resolveWorkDir(path)
	if err != nil {
		return err
	}
	err = tui.RunConfigure(tui.ConfigureOptions{
		OutputPath: path,
		Init:       true,
		WorkDir:    workDir,
	})
	if errors.Is(err, tui.ErrAborted) {
		return nil
	}
	return err
}

func parseFocus(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func printServiceList(cfg *config.Config, workDir string) {
	if _, err := cfg.SortedServiceIDs(); err != nil {
		fmt.Fprintf(os.Stderr, "muxdev: %v\n", err)
		os.Exit(1)
	}

	width := 80
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 {
		width = w
	}

	fmt.Println(tui.RenderServiceList(cfg, workDir, width))
}

