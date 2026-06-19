package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/yarkingulacti/muxdev-cli/internal/update"
	"github.com/yarkingulacti/muxdev-cli/internal/version"
)

func newUpdateCmd() *cobra.Command {
	var (
		checkOnly    bool
		yes          bool
		target       string
		channel      string
		applyPending bool
		updateURL    string
	)

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Check for and apply muxdev updates",
		RunE: func(cmd *cobra.Command, args []string) error {
			if applyPending {
				return update.ApplyPending()
			}

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()

			result, err := update.Check(ctx, update.CheckOptions{
				Channel:     update.Channel(channel),
				Target:      target,
				ManifestURL: updateURL,
			})
			if err != nil {
				return err
			}

			_ = update.WriteCache(update.CacheEntry{
				CheckedAt:       time.Now(),
				Current:         result.Current,
				Latest:          result.Latest,
				UpdateAvailable: result.UpdateAvailable,
			})

			printUpdateResult(result)

			if !result.UpdateAvailable {
				return nil
			}

			if !result.InstallMethod.SupportsSelfUpdate() {
				fmt.Printf("Installed via %s. Run: %s\n", result.InstallMethod, result.InstallMethod.UpgradeHint(result.Latest))
				if checkOnly && !isTerminal(os.Stdin) {
					return exitError{code: 2, msg: "use package manager to update"}
				}
				return nil
			}

			if version.IsDev() {
				fmt.Printf("Dev build detected. Run: %s\n", update.MethodDev.UpgradeHint(result.Latest))
				return nil
			}

			interactive := isTerminal(os.Stdin) && !yes
			if shouldExitCheckOnly(checkOnly, yes, interactive) {
				return exitError{code: 2, msg: "update available"}
			}

			if !yes && !confirmUpdate(result.Latest) {
				return nil
			}

			if err := applyUpdate(ctx, result); err != nil {
				return err
			}

			fmt.Printf("Updated to %s\n", result.Latest)
			return nil
		},
	}

	cmd.Flags().BoolVar(&checkOnly, "check", false, "Check for updates (prompts to install when interactive)")
	cmd.Flags().BoolVar(&yes, "yes", false, "Apply without confirmation")
	cmd.Flags().StringVar(&target, "version", "", "Target version tag (e.g. v0.2.0)")
	cmd.Flags().StringVar(&channel, "channel", string(update.ChannelStable), "Release channel: stable or prerelease")
	cmd.Flags().BoolVar(&applyPending, "apply-pending", false, "Apply a staged Windows update")
	cmd.Flags().StringVar(&updateURL, "update-url", "", "Manifest URL (overrides MUXDEV_UPDATE_URL)")

	return cmd
}

func printUpdateResult(result update.Result) {
	fmt.Printf("muxdev %s (installed via %s)\n", strings.TrimPrefix(result.Current, "v"), result.InstallMethod)
	if result.ManifestURL != "" {
		fmt.Printf("Update source: %s\n", result.ManifestURL)
	}
	if result.UpdateAvailable {
		fmt.Printf("Update available: %s\n", result.Latest)
	} else {
		fmt.Println("Up to date.")
	}
}

func shouldExitCheckOnly(checkOnly, yes, interactive bool) bool {
	return checkOnly && !yes && !interactive
}

func applyUpdate(ctx context.Context, result update.Result) error {
	goos, goarch := update.CurrentPlatform()
	assetName := update.PlatformAssetName(result.Latest, goos, goarch)
	return update.Apply(ctx, update.ApplyOptions{
		Release:   result.Release,
		AssetName: assetName,
	})
}

func confirmUpdate(latest string) bool {
	return confirm(confirmUpdatePrompt(latest))
}

func confirmUpdatePrompt(latest string) string {
	return fmt.Sprintf("A new update is available (%s). Would you like to install it?", latest)
}

func confirm(prompt string) bool {
	fmt.Printf("%s [y/N] ", prompt)
	reader := bufio.NewReader(os.Stdin)
	text, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	text = strings.TrimSpace(strings.ToLower(text))
	return text == "y" || text == "yes"
}
