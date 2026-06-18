package cli

import (
	"errors"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/yarkingulacti/muxdev-cli/internal/config"
	"github.com/yarkingulacti/muxdev-cli/internal/tui"
)

func newInitCmd() *cobra.Command {
	var (
		output string
		force  bool
	)

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create muxdev.yaml interactively",
		Long:  "Run an interactive wizard to generate a project muxdev.yaml manifest.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !term.IsTerminal(int(os.Stdout.Fd())) {
				return errors.New("init requires an interactive terminal; use a TTY")
			}
			path := output
			if path == "" {
				path = config.DefaultFilename
			}
			err := tui.RunConfigure(tui.ConfigureOptions{
				OutputPath: path,
				Force:      force,
				WorkDir:    ".",
			})
			if errors.Is(err, tui.ErrAborted) {
				return nil
			}
			return err
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "Output path (default: ./muxdev.yaml)")
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing file")

	return cmd
}

func newConfigureCmd() *cobra.Command {
	var (
		output string
		force  bool
	)

	cmd := &cobra.Command{
		Use:     "configure",
		Aliases: []string{"config"},
		Short:   "Edit muxdev.yaml interactively",
		Long:    "Load an existing muxdev.yaml (or create a new one) and edit it with the interactive wizard.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !term.IsTerminal(int(os.Stdout.Fd())) {
				return errors.New("configure requires an interactive terminal; use a TTY")
			}
			path := output
			if path == "" {
				cwd, err := os.Getwd()
				if err != nil {
					return err
				}
				found, err := config.FindDefault(cwd)
				if err != nil {
					path = config.DefaultFilename
				} else {
					path = found
				}
			}
			err := tui.RunConfigure(tui.ConfigureOptions{
				OutputPath: path,
				Force:      force,
				Edit:       true,
				WorkDir:    ".",
			})
			if errors.Is(err, tui.ErrAborted) {
				return nil
			}
			return err
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "Config path (default: search cwd for muxdev.yaml)")
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite without merge when creating new file")

	return cmd
}
