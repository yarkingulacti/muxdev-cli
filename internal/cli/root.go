package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/yarkingulacti/muxdev-cli/internal/config"
	"github.com/yarkingulacti/muxdev-cli/internal/runner"
	"github.com/yarkingulacti/muxdev-cli/internal/tui"
)

type Options struct {
	Version       string
	ConfigPath    string
	List          bool
	NoInteractive bool
	Focus         string
}

func NewRoot(opts Options) *cobra.Command {
	var focus string

	cmd := &cobra.Command{
		Use:   "muxdev",
		Short: "Multiplexed dev stack runner",
		Long:  "Config-driven local development orchestrator with an interactive terminal UI.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(opts, focus)
		},
	}

	cmd.Flags().StringVarP(&opts.ConfigPath, "config", "c", "", "Path to muxdev.yaml (default: search upward from cwd)")
	cmd.Flags().BoolVar(&opts.List, "list", false, "List configured services and exit")
	cmd.Flags().BoolVar(&opts.NoInteractive, "no-interactive", false, "Run without the interactive TUI")
	cmd.Flags().StringVar(&focus, "focus", "", "Comma-separated service IDs to run")

	cmd.Version = opts.Version

	return cmd
}

func run(opts Options, focus string) error {
	cfgPath, err := resolveConfigPath(opts.ConfigPath)
	if err != nil {
		return err
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
		err := tui.Run(tui.Options{
			Cfg:     cfg,
			Focus:   focusIDs,
			WorkDir: ".",
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
	return config.FindDefault(cwd)
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
