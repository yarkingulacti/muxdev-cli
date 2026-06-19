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
		if opts.List {
			return fmt.Errorf("%s not found", cfgPath)
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

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return err
	}

	if opts.List {
		printServiceList(cfg)
		return nil
	}

	focusIDs := parseFocus(focus)
	interactive := !opts.NoInteractive && isTerminal(os.Stdout)

	if interactive {
		hint := ""
		if update.ShouldCheckOnStart(24 * time.Hour) {
			hint = update.StartupHint()
			go backgroundUpdateCheck()
		}
		err := tui.Run(tui.Options{
			Cfg:        cfg,
			Focus:      focusIDs,
			WorkDir:    ".",
			UpdateHint: hint,
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

	r := runner.New(cfg, serviceIDs)
	return r.Run(cmdContext())
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
	return filepath.Join(cwd, config.DefaultFilename), nil
}

func runInitWizard(path string) error {
	err := tui.RunConfigure(tui.ConfigureOptions{
		OutputPath: path,
		Init:       true,
		WorkDir:    ".",
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

func printServiceList(cfg *config.Config) {
	ids, err := cfg.SortedServiceIDs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "muxdev: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("%s", cfg.Name)
	if cfg.Subtitle != "" {
		fmt.Printf(" — %s", cfg.Subtitle)
	}
	fmt.Println()

	for _, id := range ids {
		svc := cfg.Services[id]
		fmt.Printf("  %s (%s)\n", id, svc.Label)
		fmt.Printf("    command: %s\n", svc.Command)
		if svc.Port != "" {
			fmt.Printf("    port: %s\n", svc.Port)
		}
		if len(svc.DependsOn) > 0 {
			fmt.Printf("    depends_on: %s\n", strings.Join(svc.DependsOn, ", "))
		}
	}
}

func cmdContext() runner.Context {
	return runner.Context{
		WorkDir: ".",
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
	}
}
