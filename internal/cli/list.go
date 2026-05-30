package cli

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"

	"github.com/ahmadraza100/dotlock/internal/config"
	"github.com/ahmadraza100/dotlock/internal/crypto"
	"github.com/ahmadraza100/dotlock/internal/store"
	"github.com/ahmadraza100/dotlock/internal/ui"
)

var showValues bool

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list secrets in the active profile",
	RunE: func(_ *cobra.Command, _ []string) error {
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

		entries, err := store.ListEntries(&vault, cfg.DefaultProfile)
		if err != nil {
			ui.Error(fmt.Sprintf("profile %q not found", cfg.DefaultProfile))
			return errSilent
		}

		sort.Strings(entries)
		ui.ProfileHeader(cfg.DefaultProfile, len(entries))

		if len(entries) == 0 {
			ui.Info("no secrets yet  —  run: dotlock set KEY")
			ui.Blank()
			return nil
		}

		ui.Rule()
		for _, k := range entries {
			if showValues {
				v, err := store.GetEntry(&vault, cfg.DefaultProfile, k, ident)
				if err != nil {
					ui.Error(fmt.Sprintf("cannot decrypt %s: %v", k, err))
					return errSilent
				}
				ui.SecretRowValue(k, string(v))
				crypto.ZeroBytes(v)
			} else {
				ui.SecretRow(k)
			}
		}
		ui.Rule()
		ui.Blank()
		return nil
	},
}

func init() {
	listCmd.Flags().BoolVar(&showValues, "show-values", false, "reveal secret values")
}
