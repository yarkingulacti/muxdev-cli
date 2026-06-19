package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yarkingulacti/muxdev-cli/internal/config"
	"github.com/yarkingulacti/muxdev-cli/internal/logs"
)

func newRemoveCmd() *cobra.Command {
	var configPath string
	var yes bool

	cmd := &cobra.Command{
		Use:     "remove",
		Aliases: []string{"rm"},
		Short:   "Remove muxdev data from this project",
		Long: `Delete muxdev.yaml (or the manifest passed with --config) and all saved
runtime session logs associated with the project directory.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRemove(configPath, yes)
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "", "Path to muxdev.yaml (default: search upward from cwd)")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Remove without confirmation")

	return cmd
}

func runRemove(explicitConfig string, yes bool) error {
	cfgPath, workDir, err := resolveRemoveTargets(explicitConfig)
	if err != nil {
		return err
	}

	hasConfig := config.Exists(cfgPath)
	sessions, err := logs.ListSessions(workDir)
	if err != nil {
		return err
	}
	sessionCount := len(sessions)

	if !hasConfig && sessionCount == 0 {
		return fmt.Errorf("no muxdev data found for %s", workDir)
	}

	if !yes {
		if !isTerminal(os.Stdin) {
			return fmt.Errorf("non-interactive mode: pass --yes to confirm removal")
		}
		if !confirm(removeConfirmPrompt(cfgPath, hasConfig, sessionCount)) {
			return nil
		}
	}

	if hasConfig {
		if err := os.Remove(cfgPath); err != nil {
			return fmt.Errorf("remove config: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Removed %s\n", cfgPath)
	}

	removed, err := logs.RemoveSessions(workDir)
	if err != nil {
		return fmt.Errorf("remove sessions: %w", err)
	}
	if removed > 0 {
		fmt.Fprintf(os.Stderr, "Removed %d session log(s) for %s\n", removed, workDir)
	}

	return nil
}

func resolveRemoveTargets(explicit string) (cfgPath, workDir string, err error) {
	if explicit != "" {
		cfgPath = explicit
	} else {
		cwd, err := os.Getwd()
		if err != nil {
			return "", "", fmt.Errorf("get working directory: %w", err)
		}
		if path, err := config.FindDefault(cwd); err == nil {
			cfgPath = path
		} else {
			cfgPath = filepath.Join(cwd, config.DefaultFilename)
		}
	}

	workDir, err = resolveWorkDir(cfgPath)
	if err != nil {
		return "", "", err
	}
	return cfgPath, workDir, nil
}

func removeConfirmPrompt(cfgPath string, hasConfig bool, sessionCount int) string {
	var parts []string
	if hasConfig {
		parts = append(parts, cfgPath)
	}
	if sessionCount > 0 {
		label := "session log"
		if sessionCount != 1 {
			label += "s"
		}
		parts = append(parts, fmt.Sprintf("%d %s", sessionCount, label))
	}
	return fmt.Sprintf("Remove %s from this project?", strings.Join(parts, " and "))
}
