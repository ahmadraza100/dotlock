package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/ahmadraza100/dotlock/internal/config"
	"github.com/ahmadraza100/dotlock/internal/crypto"
	"github.com/ahmadraza100/dotlock/internal/store"
	"github.com/ahmadraza100/dotlock/internal/ui"
)

var getCmd = &cobra.Command{
	Use:   "get KEY",
	Short: "print a secret value to stdout",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return RunInteractiveGet(cmd)
		}
		key := args[0]
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

		value, err := store.GetEntry(&vault, cfg.DefaultProfile, key, ident)
		if err != nil {
			ui.Error(fmt.Sprintf("%s not found in profile %q", key, cfg.DefaultProfile))
			return errSilent
		}

		os.Stdout.Write(value)
		if len(value) > 0 && value[len(value)-1] != '\n' {
			fmt.Println()
		}
		return nil
	},
}
