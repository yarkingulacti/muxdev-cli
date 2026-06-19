package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/yarkingulacti/muxdev-cli/internal/config"
)

func newListCmd() *cobra.Command {
	var configPath string

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List configured services",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(configPath)
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "", "Path to muxdev.yaml (default: search upward from cwd)")

	return cmd
}

func runList(explicitConfig string) error {
	cfgPath, err := resolveConfigPath(explicitConfig)
	if err != nil {
		return err
	}
	if !config.Exists(cfgPath) {
		return fmt.Errorf("%s not found", cfgPath)
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return err
	}
	workDir, err := resolveWorkDir(cfgPath)
	if err != nil {
		return err
	}
	printServiceList(cfg, workDir)
	return nil
}
