package cli

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/ahmadraza100/dotlock/internal/config"
	"github.com/ahmadraza100/dotlock/internal/crypto"
	"github.com/ahmadraza100/dotlock/internal/model"
	"github.com/ahmadraza100/dotlock/internal/store"
	"github.com/ahmadraza100/dotlock/internal/ui"
)

var exportFormat string

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "export the active profile  (--format shell|docker|github)",
	RunE: func(cmd *cobra.Command, _ []string) error {
		if !cmd.Flags().Changed("format") {
			return RunInteractiveExport(cmd)
		}
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

		profileMap, err := store.ProfileMap(&vault, cfg.DefaultProfile, ident)
		if err != nil {
			ui.Error(fmt.Sprintf("profile %q not found", cfg.DefaultProfile))
			return errSilent
		}

		switch exportFormat {
		case "shell":
			fmt.Println(model.RenderShell(profileMap))
		case "docker":
			fmt.Println(model.RenderDocker(profileMap))
		case "github":
			fmt.Println(model.RenderGitHubActions(profileMap))
		default:
			ui.Error(fmt.Sprintf("unknown format %q  (shell | docker | github)", exportFormat))
			return errSilent
		}

		for _, v := range profileMap {
			crypto.ZeroBytes(v)
		}
		return nil
	},
}

func init() {
	exportCmd.Flags().StringVar(&exportFormat, "format", "shell", "output format: shell | docker | github")
}
