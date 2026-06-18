package main

import (
	"fmt"
	"os"

	"github.com/yarkingulacti/muxdev-cli/internal/cli"
)

func main() {
	root := cli.NewRoot()
	if err := root.Execute(); err != nil {
		if code := cli.ExitCode(err); code != 0 {
			if code > 0 {
				fmt.Fprintf(os.Stderr, "muxdev: %v\n", err)
			}
			os.Exit(code)
		}
		fmt.Fprintf(os.Stderr, "muxdev: %v\n", err)
		os.Exit(1)
	}
}
