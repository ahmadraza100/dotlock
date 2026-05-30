package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/ahmadraza100/dotlock/internal/config"
	"github.com/ahmadraza100/dotlock/internal/crypto"
	"github.com/ahmadraza100/dotlock/internal/store"
	"github.com/ahmadraza100/dotlock/internal/ui"
)

var initYes bool

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "create a new encrypted vault in the current directory",
	RunE: func(cmd *cobra.Command, _ []string) error {
		if !initYes {
			return RunInteractiveInit(cmd)
		}
		return createVault(true)
	},
}

// createVault creates a vault in the current directory with defaults.
// If printDetails is true it prints the post-init summary; used by `init -y`.
// Called silently by the root command when no vault exists yet.
func createVault(printDetails bool) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("cannot determine current directory: %w", err)
	}

	cfg := config.LoadConfig()
	vaultPath := filepath.Join(cwd, cfg.VaultFilename)

	recipient, _, err := crypto.LoadDefaultRecipientAndIdentity()
	if err != nil {
		return fmt.Errorf("cannot load or create identity: %w", err)
	}

	vault := store.NewEmptyVault()
	vault.ID = uuid.New()
	vault.Name = filepath.Base(cwd)
	vault.CreatedAt = time.Now().UTC()
	vault.UpdatedAt = vault.CreatedAt
	vault.Profiles = map[string]store.Profile{"dev": {Entries: map[string]store.Entry{}}}

	data, err := store.MarshalAndEncryptVault(&vault, recipient)
	if err != nil {
		return fmt.Errorf("cannot encrypt vault: %w", err)
	}
	if err := store.AtomicWrite(vaultPath, data, 0600); err != nil {
		return fmt.Errorf("cannot write vault: %w", err)
	}

	if printDetails {
		fmt.Println()
		ui.Success(fmt.Sprintf("vault created  %s", ui.Highlight(cfg.VaultFilename)))
		fmt.Println()
		ui.KV("default profile", "dev")
		ui.KV("vault id", vault.ID.String())
		fmt.Println()
		ui.KV("next steps", "dotlock set DATABASE_URL")
		ui.KV("", "dotlock list")
		ui.KV("", "dotlock profile create staging")
		fmt.Println()
	}
	return nil
}

// ensureVault creates a vault in the current directory if one does not exist.
func ensureVault() error {
	cfg := config.LoadConfig()
	vaultPath := filepath.Join(".", cfg.VaultFilename)
	if _, err := os.Stat(vaultPath); err == nil {
		return nil // already exists
	}
	return createVault(false)
}

func init() {
	initCmd.Flags().BoolVarP(&initYes, "yes", "y", false, "skip interactive prompts and create vault with defaults")
}
