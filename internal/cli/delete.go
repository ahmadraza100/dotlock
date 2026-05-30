package cli

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/ahmadraza100/dotlock/internal/config"
	"github.com/ahmadraza100/dotlock/internal/crypto"
	"github.com/ahmadraza100/dotlock/internal/store"
	"github.com/ahmadraza100/dotlock/internal/ui"
)

var deleteCmd = &cobra.Command{
	Use:   "delete KEY",
	Short: "delete a secret from the active profile",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return RunInteractiveDelete(cmd)
		}
		key := args[0]
		cfg := config.LoadConfig()
		vaultPath := filepath.Join(".", cfg.VaultFilename)

		rec, ident, err := crypto.LoadDefaultRecipientAndIdentity()
		if err != nil {
			ui.Error(fmt.Sprintf("cannot load crypto keys: %v", err))
			return errSilent
		}

		vault, err := store.LoadVault(vaultPath, ident)
		if err != nil {
			ui.Error(fmt.Sprintf("cannot load vault: %v", err))
			return errSilent
		}

		if err := store.DeleteEntry(&vault, cfg.DefaultProfile, key); err != nil {
			ui.Error(fmt.Sprintf("%s not found in profile %q", key, cfg.DefaultProfile))
			return errSilent
		}

		data, err := store.MarshalAndEncryptVault(&vault, rec)
		if err != nil {
			ui.Error(fmt.Sprintf("cannot encrypt vault: %v", err))
			return errSilent
		}
		if err := store.AtomicWrite(vaultPath, data, 0600); err != nil {
			ui.Error(fmt.Sprintf("cannot write vault: %v", err))
			return errSilent
		}

		ui.Success(fmt.Sprintf("%s  deleted from  %s", ui.Highlight(key), ui.Dim(cfg.DefaultProfile)))
		return nil
	},
}
