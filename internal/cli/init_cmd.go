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
		Short: "Create muxdev.yaml with an interactive setup wizard",
		Long: `Create a muxdev.yaml manifest for your project using an interactive wizard.

The wizard walks you through project metadata, dev services (commands, ports,
dependencies), and shows a preview before writing the file.

Run from your project root:

  muxdev init`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !term.IsTerminal(int(os.Stdout.Fd())) {
				return errors.New("init requires an interactive terminal; use a TTY")
			}
			path := output
			if path == "" {
				path = config.DefaultFilename
			}
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			err = tui.RunConfigure(tui.ConfigureOptions{
				OutputPath: path,
				Force:      force,
				Init:       true,
				WorkDir:    cwd,
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
		Long:    "Load muxdev.yaml and edit only the fields you choose — no step-by-step wizard.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !term.IsTerminal(int(os.Stdout.Fd())) {
				return errors.New("configure requires an interactive terminal; use a TTY")
			}
			path := output
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			if path == "" {
				found, err := config.FindDefault(cwd)
				if err != nil {
					path = config.DefaultFilename
				} else {
					path = found
				}
			}
			err = tui.RunConfigure(tui.ConfigureOptions{
				OutputPath: path,
				Force:      force,
				Edit:       true,
				WorkDir:    cwd,
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
