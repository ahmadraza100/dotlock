package main

import (
	"context"
	"os"

	"github.com/ahmadraza100/dotlock/internal/cli"
	"github.com/ahmadraza100/dotlock/internal/ui"
)

func main() {
	root := cli.RootCommand()
	if err := root.ExecuteContext(context.Background()); err != nil {
		// cobra validation errors (wrong arg count, unknown flag) have a non-empty
		// message and were not printed by the command itself.
		if msg := err.Error(); msg != "" {
			ui.Error(msg)
		}
		os.Exit(1)
	}
}
