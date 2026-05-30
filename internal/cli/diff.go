package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ahmadraza100/dotlock/internal/config"
	"github.com/ahmadraza100/dotlock/internal/crypto"
	"github.com/ahmadraza100/dotlock/internal/store"
	"github.com/ahmadraza100/dotlock/internal/ui"
	"github.com/ahmadraza100/dotlock/pkg/diff"
)

var diffCmd = &cobra.Command{
	Use:   "diff PROFILE_A PROFILE_B",
	Short: "compare two profiles",
	Args:  cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return RunInteractiveDiff(cmd)
		}
		a, b := args[0], args[1]
		cfg := config.LoadConfig()
		vaultPath := filepath.Join(".", cfg.VaultFilename)

		_, ident, err := crypto.LoadDefaultRecipientAndIdentity()
		if err != nil {
			ui.Error(fmt.Sprintf("cannot load identity: %v", err))
			return errSilent
		}

		vault, err := store.LoadVault(vaultPath, ident)
		if err != nil {
			ui.Error(fmt.Sprintf("cannot load vault: %v", err))
			return errSilent
		}

		da, err := store.ProfileMap(&vault, a, ident)
		if err != nil {
			ui.Error(fmt.Sprintf("profile %q not found", a))
			return errSilent
		}
		db, err := store.ProfileMap(&vault, b, ident)
		if err != nil {
			ui.Error(fmt.Sprintf("profile %q not found", b))
			return errSilent
		}

		lines := diff.Maps(da, db)

		fmt.Println()
		fmt.Printf("  %s  %s  %s\n", ui.Highlight(a), ui.Dim("→"), ui.Highlight(b))
		fmt.Println()

		if len(lines) == 0 {
			ui.Info("no keys in either profile")
			fmt.Println()
			return nil
		}

		ui.Rule()
		for _, line := range lines {
			if len(line) < 2 {
				continue
			}
			symbol := strings.TrimSpace(line[:1])
			key := strings.TrimSpace(line[2:])
			ui.DiffLine(symbol, key)
		}
		ui.Rule()
		fmt.Println()

		for _, v := range da {
			crypto.ZeroBytes(v)
		}
		for _, v := range db {
			crypto.ZeroBytes(v)
		}
		return nil
	},
}
