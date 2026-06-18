package main

import (
	"fmt"
	"os"

	"github.com/yarkingulacti/muxdev-cli/internal/cli"
)

var version = "0.1.0"

func main() {
	root := cli.NewRoot(cli.Options{Version: version})
	if err := root.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "muxdev: %v\n", err)
		os.Exit(1)
	}
}
