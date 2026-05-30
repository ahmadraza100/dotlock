package cli

import (
	"fmt"
	"os"

	"log/slog"

	"github.com/spf13/cobra"

	"github.com/ahmadraza100/dotlock/internal/ui"
)

var (
	version = "v0.1.0"
	debug   = false
)

// RootCommand returns the root cobra command.
func RootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:           "dotlock",
		Short:         "dotlock — encrypted .env vaults across profiles and projects",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := ensureVault(); err != nil {
				ui.Error(err.Error())
				return errSilent
			}
			return uiCmd.RunE(cmd, nil)
		},
	}

	root.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug logging")
	root.AddCommand(initCmd)
	root.AddCommand(setCmd)
	root.AddCommand(getCmd)
	root.AddCommand(deleteCmd)
	root.AddCommand(listCmd)
	root.AddCommand(diffCmd)
	root.AddCommand(exportCmd)
	root.AddCommand(profileCmd)
	root.AddCommand(uiCmd)
	root.AddCommand(versionCmd)

	cobra.OnInitialize(func() {
		level := slog.LevelWarn
		if debug {
			level = slog.LevelDebug
		}
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})))
	})

	return root
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "print version",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf("dotlock  %s\n", version)
	},
}
