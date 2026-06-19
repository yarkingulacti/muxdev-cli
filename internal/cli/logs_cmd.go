package cli

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/yarkingulacti/muxdev-cli/internal/logs"
	"github.com/yarkingulacti/muxdev-cli/internal/platform"
	"github.com/yarkingulacti/muxdev-cli/internal/tui"
)

func newLogsCmd() *cobra.Command {
	var configPath string
	var allProjects bool

	cmd := &cobra.Command{
		Use:   "logs",
		Short: "Browse persisted runtime session logs",
		Long:  "Runtime sessions are stored under the platform-specific muxdev sessions directory.",
		RunE: func(cmd *cobra.Command, args []string) error {
			workDir, err := resolveLogsWorkDir(configPath, allProjects)
			if err != nil {
				return err
			}
			if isTerminal(os.Stdout) {
				return tui.RunSessionLogs(tui.SessionLogsOptions{
					WorkDir: workDir,
					All:     allProjects,
				})
			}
			return runLogsList(workDir, allProjects)
		},
	}

	cmd.AddCommand(newLogsListCmd())
	cmd.AddCommand(newLogsShowCmd())
	cmd.AddCommand(newLogsPathCmd())

	cmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "Path to muxdev.yaml (default: search upward from cwd)")
	cmd.PersistentFlags().BoolVar(&allProjects, "all", false, "Include sessions from all projects")

	return cmd
}

func newLogsListCmd() *cobra.Command {
	var configPath string
	var allProjects bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List saved runtime sessions",
		RunE: func(cmd *cobra.Command, args []string) error {
			workDir, err := resolveLogsWorkDir(configPath, allProjects)
			if err != nil {
				return err
			}
			return runLogsList(workDir, allProjects)
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "", "Path to muxdev.yaml (default: search upward from cwd)")
	cmd.Flags().BoolVar(&allProjects, "all", false, "Include sessions from all projects")

	return cmd
}

func newLogsShowCmd() *cobra.Command {
	var configPath string
	var allProjects bool
	var latest bool

	cmd := &cobra.Command{
		Use:   "show [session-id]",
		Short: "Print a saved session log",
		RunE: func(cmd *cobra.Command, args []string) error {
			workDir, err := resolveLogsWorkDir(configPath, allProjects)
			if err != nil {
				return err
			}
			return runLogsShow(workDir, allProjects, latest, args)
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "", "Path to muxdev.yaml (default: search upward from cwd)")
	cmd.Flags().BoolVar(&allProjects, "all", false, "Include sessions from all projects")
	cmd.Flags().BoolVar(&latest, "latest", false, "Show the most recent session")

	return cmd
}

func newLogsPathCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Print the platform-specific sessions directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := platform.SessionsDir()
			if err != nil {
				return err
			}
			fmt.Println(dir)
			return nil
		},
	}
}

func resolveLogsWorkDir(configPath string, allProjects bool) (string, error) {
	if allProjects {
		return "", nil
	}
	cfgPath, err := resolveConfigPath(configPath)
	if err != nil {
		return "", err
	}
	return resolveWorkDir(cfgPath)
}

func runLogsList(workDir string, allProjects bool) error {
	sessions, err := logs.ListSessions(workDir)
	if err != nil {
		return err
	}
	if len(sessions) == 0 {
		scope := "this project"
		if allProjects {
			scope = "any project"
		}
		fmt.Printf("No saved runtime sessions for %s.\n", scope)
		return nil
	}

	fmt.Printf("%-20s  %-19s  %-8s  %s\n", "SESSION", "STARTED", "RUNTIME", "SERVICES")
	for _, session := range sessions {
		status := "running"
		if session.Meta.EndedAt != nil {
			status = "done"
			if session.Meta.ExitError != "" {
				status = "error"
			}
		}
		started := session.Meta.StartedAt.Local().Format("2006-01-02 15:04:05")
		services := strings.Join(session.Meta.ServiceIDs, ",")
		fmt.Printf("%-20s  %-19s  %-8s  %s (%s)\n",
			session.Meta.ID,
			started,
			session.Meta.Runtime,
			services,
			status,
		)
	}
	return nil
}

func runLogsShow(workDir string, allProjects bool, latest bool, args []string) error {
	sessions, err := logs.ListSessions(workDir)
	if err != nil {
		return err
	}
	if len(sessions) == 0 {
		return fmt.Errorf("no saved runtime sessions found")
	}

	var session logs.Session
	switch {
	case latest:
		session = sessions[0]
	case len(args) == 1:
		id := args[0]
		found := false
		for _, candidate := range sessions {
			if candidate.Meta.ID == id || strings.HasPrefix(candidate.Meta.ID, id) {
				session = candidate
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("session %q not found", id)
		}
	default:
		return fmt.Errorf("session id required (or use --latest)")
	}

	content, err := logs.ReadLog(session.Dir)
	if err != nil {
		return err
	}
	if content == "" {
		fmt.Printf("# session %s (%s)\n", session.Meta.ID, session.Meta.StartedAt.Local().Format(time.RFC3339))
		return nil
	}
	fmt.Print(content)
	return nil
}
