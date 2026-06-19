package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yarkingulacti/muxdev-cli/internal/cli/doc"
)

func newHelpCmd(root *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "help [topic]",
		Short: "Interactive local guide and command reference",
		Long:  "Browse muxdev documentation in the terminal. Topics are generated from registered commands and stay in sync with the CLI.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runHelp(root, cmd.OutOrStdout(), strings.Join(args, " "))
		},
	}
	return cmd
}

func printTopic(out io.Writer, pages []doc.Page, query string) error {
	page, ok := doc.FindPage(pages, query)
	if !ok {
		return fmt.Errorf("unknown help topic %q — run muxdev help for the index", query)
	}
	fmt.Fprintf(out, "%s\n\n%s\n", page.Title, page.Body)
	if page.TryCommand != "" {
		fmt.Fprintf(out, "\nTry: %s\n", page.TryCommand)
	}
	return nil
}

func registerHelp(root *cobra.Command) {
	doc.Attach(root)
	helpCmd := newHelpCmd(root)
	root.AddCommand(helpCmd)
	root.SetHelpCommand(helpCmd)
	disableHelpFlag(root)
}