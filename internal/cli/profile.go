package cli

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sort"

	"github.com/spf13/cobra"

	"github.com/ahmadraza100/dotlock/internal/config"
	"github.com/ahmadraza100/dotlock/internal/crypto"
	"github.com/ahmadraza100/dotlock/internal/store"
	"github.com/ahmadraza100/dotlock/internal/ui"
)

var profileNameRegexp = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "manage profiles  (create | list | use | delete)",
	RunE: func(cmd *cobra.Command, _ []string) error {
		return RunInteractiveProfile(cmd)
	},
}

var profileCreateCmd = &cobra.Command{
	Use:   "create NAME",
	Short: "create a new profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		name := args[0]
		if !profileNameRegexp.MatchString(name) {
			ui.Error(fmt.Sprintf("invalid name %q  (must match [a-z][a-z0-9-]*)", name))
			return errSilent
		}

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

		if err := store.CreateProfile(&vault, name); err != nil {
			ui.Error(fmt.Sprintf("cannot create profile: %v", err))
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

		ui.Success(fmt.Sprintf("profile  %s  created", ui.Highlight(name)))
		return nil
	},
}

var profileListCmd = &cobra.Command{
	Use:   "list",
	Short: "list all profiles",
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

		names := make([]string, 0, len(vault.Profiles))
		for n := range vault.Profiles {
			names = append(names, n)
		}
		sort.Strings(names)

		fmt.Println()
		fmt.Printf("  %s  %s  %s  %s\n",
			ui.Dim(fmt.Sprintf("%d profiles", len(names))),
			ui.Dim("·"),
			ui.Dim("active:"),
			ui.Highlight(cfg.DefaultProfile),
		)
		fmt.Println()
		for _, n := range names {
			ui.ProfileItem(n, n == cfg.DefaultProfile)
		}
		fmt.Println()
		return nil
	},
}

var profileUseCmd = &cobra.Command{
	Use:   "use NAME",
	Short: "switch the active profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		name := args[0]
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

		if _, ok := vault.Profiles[name]; !ok {
			ui.Error(fmt.Sprintf("profile %q does not exist", name))
			return errSilent
		}

		cfg.DefaultProfile = name
		if err := config.SaveConfig(cfg); err != nil {
			ui.Error(fmt.Sprintf("cannot save config: %v", err))
			return errSilent
		}

		ui.Success(fmt.Sprintf("switched to  %s", ui.Highlight(name)))
		return nil
	},
}

var profileDeleteCmd = &cobra.Command{
	Use:   "delete NAME",
	Short: "delete a profile and all its secrets",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		name := args[0]
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

		if err := store.DeleteProfile(&vault, name); err != nil {
			ui.Error(fmt.Sprintf("cannot delete profile: %v", err))
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

		ui.Success(fmt.Sprintf("profile  %s  deleted", ui.Highlight(name)))
		return nil
	},
}

func init() {
	profileCmd.AddCommand(profileCreateCmd)
	profileCmd.AddCommand(profileListCmd)
	profileCmd.AddCommand(profileUseCmd)
	profileCmd.AddCommand(profileDeleteCmd)
}
