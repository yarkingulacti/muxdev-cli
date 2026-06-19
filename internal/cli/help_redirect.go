package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yarkingulacti/muxdev-cli/internal/cli/doc"
	"github.com/yarkingulacti/muxdev-cli/internal/tui"
)

func runHelp(root *cobra.Command, out io.Writer, topic string) error {
	pages := doc.BuildPages(root)
	topic = strings.TrimSpace(topic)

	if isTerminal(os.Stdout) {
		opts := tui.WikiOptions{Pages: pages}
		if topic != "" {
			opts.OpenTopic = topic
		}
		return tui.RunWiki(opts)
	}

	if topic != "" {
		return printTopic(out, pages, topic)
	}
	fmt.Fprint(out, doc.IndexPlain(pages))
	return nil
}

func disableHelpFlag(root *cobra.Command) {
	applyHelpHandler(root, func(c *cobra.Command) {
		topic := helpTopicFor(c)
		if err := runHelp(root, c.OutOrStdout(), topic); err != nil {
			fmt.Fprintf(c.ErrOrStderr(), "muxdev: %v\n", err)
			os.Exit(1)
		}
	})
}

func applyHelpHandler(cmd *cobra.Command, fn func(*cobra.Command)) {
	if cmd == nil {
		return
	}
	if cmd.Name() != "help" {
		if cmd.Flags().Lookup("help") == nil {
			cmd.Flags().BoolP("help", "h", false, "")
		}
		_ = cmd.Flags().MarkHidden("help")
		cmd.SetHelpFunc(func(c *cobra.Command, _ []string) {
			fn(c)
		})
	}
	for _, sub := range cmd.Commands() {
		applyHelpHandler(sub, fn)
	}
}

func helpTopicFor(c *cobra.Command) string {
	return strings.TrimSpace(strings.TrimPrefix(c.CommandPath(), "muxdev"))
}
