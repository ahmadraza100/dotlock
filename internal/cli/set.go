package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ahmadraza100/dotlock/internal/config"
	"github.com/ahmadraza100/dotlock/internal/crypto"
	"github.com/ahmadraza100/dotlock/internal/store"
	"github.com/ahmadraza100/dotlock/internal/ui"
)

var keyNameRegexp = regexp.MustCompile(`^[A-Z_][A-Z0-9_]*$`)

var setCmd = &cobra.Command{
	Use:   "set KEY",
	Short: "add or update a secret in the active profile",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return RunInteractiveSet(cmd)
		}
		key := args[0]
		if !keyNameRegexp.MatchString(key) {
			ui.Error(fmt.Sprintf("invalid key %q  (must match [A-Z_][A-Z0-9_]*)", key))
			return errSilent
		}

		cfg := config.LoadConfig()
		vaultPath := filepath.Join(".", cfg.VaultFilename)

		rec, ident, err := crypto.LoadDefaultRecipientAndIdentity()
		if err != nil {
			ui.Error(fmt.Sprintf("cannot load crypto keys: %v", err))
			return errSilent
		}

		// prompt goes to stderr so piped input stays clean
		fmt.Fprintf(os.Stderr, "  %s  %s\n  %s",
			ui.Dim("profile"), ui.Highlight(cfg.DefaultProfile),
			ui.Dim("value: "),
		)

		reader := bufio.NewReader(os.Stdin)
		raw, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			ui.Error(fmt.Sprintf("cannot read value: %v", err))
			return errSilent
		}
		valueBytes := []byte(strings.TrimRight(raw, "\r\n"))

		vault, err := store.LoadVault(vaultPath, ident)
		if err != nil {
			ui.Error(fmt.Sprintf("cannot load vault: %v", err))
			return errSilent
		}

		if err := store.SetEntry(&vault, cfg.DefaultProfile, key, valueBytes, rec); err != nil {
			crypto.ZeroBytes(valueBytes)
			ui.Error(fmt.Sprintf("cannot store secret: %v", err))
			return errSilent
		}
		crypto.ZeroBytes(valueBytes)

		data, err := store.MarshalAndEncryptVault(&vault, rec)
		if err != nil {
			ui.Error(fmt.Sprintf("cannot encrypt vault: %v", err))
			return errSilent
		}
		if err := store.AtomicWrite(vaultPath, data, 0600); err != nil {
			ui.Error(fmt.Sprintf("cannot write vault: %v", err))
			return errSilent
		}

		fmt.Println()
		ui.Success(fmt.Sprintf("%s  stored", ui.Highlight(key)))
		return nil
	},
}
