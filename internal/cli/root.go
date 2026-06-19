package cli

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/yarkingulacti/muxdev-cli/internal/version"
)

func newVersionCmd() *cobra.Command {
	var short bool
	var asJSON bool

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			if asJSON {
				out, err := version.JSON()
				if err != nil {
					return err
				}
				fmt.Println(out)
				return nil
			}
			if short {
				fmt.Println(version.Short())
				return nil
			}
			fmt.Println(version.String())
			return nil
		},
	}

	cmd.Flags().BoolVar(&short, "short", false, "Print version number only")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Print version as JSON")

	return cmd
}

func NewRoot() *cobra.Command {
	var focus string
	opts := defaultOptions()

	root := &cobra.Command{
		Use:     "muxdev",
		Short:   "Multiplexed dev stack runner",
		Long:    "Config-driven local development orchestrator with an interactive terminal UI.",
		Version: version.String(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(opts, focus)
		},
	}

	root.Flags().StringVarP(&opts.ConfigPath, "config", "c", "", "Path to muxdev.yaml (default: search upward from cwd)")
	root.Flags().BoolVar(&opts.NoInteractive, "no-interactive", false, "Run without the interactive TUI")
	root.Flags().StringVar(&opts.Runtime, "runtime", "", "Start mode: sync (sequential, dependency order) or async (parallel)")
	root.Flags().StringVar(&focus, "focus", "", "Comma-separated service IDs to run")

	root.AddCommand(newListCmd())
	root.AddCommand(newVersionCmd())
	root.AddCommand(newUpdateCmd())
	root.AddCommand(newInitCmd())
	root.AddCommand(newConfigureCmd())

	return root
}

func defaultOptions() Options {
	return Options{}
}

type exitError struct {
	code int
	msg  string
}

func (e exitError) Error() string {
	return e.msg
}

func ExitCode(err error) int {
	if err == nil {
		return 0
	}
	var ee exitError
	if errors.As(err, &ee) {
		return ee.code
	}
	return 1
}
