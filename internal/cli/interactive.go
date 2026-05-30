package cli

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/ahmadraza100/dotlock/internal/config"
	"github.com/ahmadraza100/dotlock/internal/crypto"
	"github.com/ahmadraza100/dotlock/internal/model"
	"github.com/ahmadraza100/dotlock/internal/store"
	"github.com/ahmadraza100/dotlock/internal/ui"
	"github.com/ahmadraza100/dotlock/pkg/diff"
)

var (
	iSuccessStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("78")).Bold(true)
	iErrorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	iWarnStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
	iDangerStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	iGreenStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("78"))
	iRedStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	iYellowStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	iGrayStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

func iPrintSuccess(msg string) { fmt.Printf("  %s  %s\n", iSuccessStyle.Render("✓"), msg) }
func iPrintError(msg string)   { fmt.Printf("  %s  %s\n", iErrorStyle.Render("✗"), msg) }
func iPrintWarn(msg string)    { fmt.Printf("  %s  %s\n", iWarnStyle.Render("!"), msg) }

var iHeaderStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("81")).
	Bold(true)

var iSubStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("240"))

var iDiffBannerStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("240")).
	Padding(0, 2)

func iMenuHeader() {
	logo := lipgloss.NewStyle().
		Foreground(lipgloss.Color("63")).
		Bold(true).
		Render("dotlock")
	tagline := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render("encrypted .env vault")
	brand := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(0, 2).
		Render(logo + "  " + tagline)
	welcome := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render("welcome back — your secrets are safe")
	fmt.Printf("\n  %s\n  %s\n\n", brand, welcome)
}

func iDiffBanner(a, b string) {
	cyan := lipgloss.NewStyle().Foreground(lipgloss.Color("81")).Bold(true)
	arrow := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("→")
	inner := fmt.Sprintf("%s  %s  %s", cyan.Render(a), arrow, cyan.Render(b))
	fmt.Printf("\n  %s\n\n", iDiffBannerStyle.Render(inner))
}

var iKeyRegexp     = regexp.MustCompile(`^[A-Z_][A-Z0-9_]*$`)
var iProfileRegexp = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

func iVaultExists() bool {
	cfg := config.LoadConfig()
	_, err := os.Stat(filepath.Join(".", cfg.VaultFilename))
	return err == nil
}

func iProfileNames(vault store.Vault) []string {
	names := make([]string, 0, len(vault.Profiles))
	for n := range vault.Profiles {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

func iProfileOptions(vault store.Vault) []huh.Option[string] {
	names := iProfileNames(vault)
	opts := make([]huh.Option[string], len(names))
	for i, n := range names {
		opts[i] = huh.NewOption(n, n)
	}
	return opts
}

func iLoadVault() (store.Vault, error) {
	cfg := config.LoadConfig()
	_, ident, err := crypto.LoadDefaultRecipientAndIdentity()
	if err != nil {
		return store.Vault{}, fmt.Errorf("cannot load crypto keys: %w", err)
	}
	vault, err := store.LoadVault(filepath.Join(".", cfg.VaultFilename), ident)
	if err != nil {
		return store.Vault{}, fmt.Errorf("cannot load vault: %w", err)
	}
	return vault, nil
}

// RunMainMenu is shown when dotlock is invoked with no subcommand.
// It loops until the user selects Quit or presses Ctrl+C.
func RunMainMenu(_ *cobra.Command) error {
	for {
		iMenuHeader()
		var choice string
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("What would you like to do?").
					Options(
						huh.NewOption("Open TUI  (visual vault manager)", "ui"),
						huh.NewOption("Add a secret", "set"),
						huh.NewOption("Get a secret", "get"),
						huh.NewOption("List secrets", "list"),
						huh.NewOption("Delete a secret", "delete"),
						huh.NewOption("Diff profiles", "diff"),
						huh.NewOption("Export secrets", "export"),
						huh.NewOption("Manage profiles", "profile"),
						huh.NewOption("Initialize vault", "init"),
						huh.NewOption("Quit", "quit"),
					).
					Value(&choice),
			),
		)
		if err := form.Run(); err != nil {
			if errors.Is(err, huh.ErrUserAborted) {
				return nil
			}
			return err
		}

		var err error
		switch choice {
		case "init":
			err = RunInteractiveInit(nil)
		case "set":
			err = RunInteractiveSet(nil)
		case "get":
			err = RunInteractiveGet(nil)
		case "list":
			err = listCmd.RunE(listCmd, nil)
		case "delete":
			err = RunInteractiveDelete(nil)
		case "diff":
			err = RunInteractiveDiff(nil)
		case "export":
			err = RunInteractiveExport(nil)
		case "profile":
			err = RunInteractiveProfile(nil)
		case "ui":
			err = uiCmd.RunE(uiCmd, nil)
		case "quit":
			return nil
		}
		if err != nil {
			return err
		}
		fmt.Println()
	}
}

// RunInteractiveSet collects key/value via prompts then stores the secret.
func RunInteractiveSet(_ *cobra.Command) error {
	if !iVaultExists() {
		iPrintError("no vault found in current directory — run dotlock init first")
		return nil
	}

	vault, err := iLoadVault()
	if err != nil {
		iPrintError(err.Error())
		return nil
	}

	cfg := config.LoadConfig()
	profiles := iProfileOptions(vault)
	if len(profiles) == 0 {
		iPrintError("no profiles found")
		return nil
	}

	selectedProfile := cfg.DefaultProfile
	var keyName string
	var secretValue string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Which profile?").
				Options(profiles...).
				Value(&selectedProfile),
			huh.NewInput().
				Title("Secret name (e.g. DATABASE_URL)").
				Validate(func(s string) error {
					if !iKeyRegexp.MatchString(s) {
						return fmt.Errorf("must be uppercase letters, numbers, and underscores only")
					}
					return nil
				}).
				Value(&keyName),
			huh.NewInput().
				Title("Secret value").
				EchoMode(huh.EchoModePassword).
				Value(&secretValue),
		),
	)
	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return nil
		}
		return err
	}

	var confirm bool
	confirmForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(fmt.Sprintf("Save %s to %s profile?", keyName, selectedProfile)).
				Affirmative("Yes").
				Negative("No").
				Value(&confirm),
		),
	)
	if err := confirmForm.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return nil
		}
		return err
	}
	if !confirm {
		return nil
	}

	valueBytes := []byte(secretValue)
	// best-effort zero of the string backing array
	secretValue = strings.Repeat("\x00", len(secretValue))
	_ = secretValue

	rec, ident, err := crypto.LoadDefaultRecipientAndIdentity()
	if err != nil {
		crypto.ZeroBytes(valueBytes)
		iPrintError(fmt.Sprintf("cannot load crypto keys: %v", err))
		return nil
	}

	freshVault, err := store.LoadVault(filepath.Join(".", cfg.VaultFilename), ident)
	if err != nil {
		crypto.ZeroBytes(valueBytes)
		iPrintError(fmt.Sprintf("cannot load vault: %v", err))
		return nil
	}

	if err := store.SetEntry(&freshVault, selectedProfile, keyName, valueBytes, rec); err != nil {
		crypto.ZeroBytes(valueBytes)
		iPrintError(fmt.Sprintf("cannot store secret: %v", err))
		return nil
	}
	crypto.ZeroBytes(valueBytes)

	data, err := store.MarshalAndEncryptVault(&freshVault, rec)
	if err != nil {
		iPrintError(fmt.Sprintf("cannot encrypt vault: %v", err))
		return nil
	}
	if err := store.AtomicWrite(filepath.Join(".", cfg.VaultFilename), data, 0600); err != nil {
		iPrintError(fmt.Sprintf("cannot write vault: %v", err))
		return nil
	}

	iPrintSuccess(fmt.Sprintf("%s saved to %s", ui.Highlight(keyName), ui.Highlight(selectedProfile)))
	return nil
}

// RunInteractiveGet lets the user pick a key and reveals/copies the value in-terminal.
func RunInteractiveGet(_ *cobra.Command) error {
	if !iVaultExists() {
		iPrintError("no vault found in current directory — run dotlock init first")
		return nil
	}

	cfg := config.LoadConfig()
	_, ident, err := crypto.LoadDefaultRecipientAndIdentity()
	if err != nil {
		iPrintError(fmt.Sprintf("cannot load crypto keys: %v", err))
		return nil
	}

	vault, err := store.LoadVault(filepath.Join(".", cfg.VaultFilename), ident)
	if err != nil {
		iPrintError(fmt.Sprintf("cannot load vault: %v", err))
		return nil
	}

	profiles := iProfileOptions(vault)
	if len(profiles) == 0 {
		iPrintError("no profiles found")
		return nil
	}

	selectedProfile := cfg.DefaultProfile
	profileForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Which profile?").
				Options(profiles...).
				Value(&selectedProfile),
		),
	)
	if err := profileForm.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return nil
		}
		return err
	}

	keys, err := store.ListEntries(&vault, selectedProfile)
	if err != nil {
		iPrintError(fmt.Sprintf("profile %q not found", selectedProfile))
		return nil
	}
	if len(keys) == 0 {
		iPrintWarn(fmt.Sprintf("no secrets in profile %q", selectedProfile))
		return nil
	}
	sort.Strings(keys)

	keyOpts := make([]huh.Option[string], len(keys))
	for i, k := range keys {
		keyOpts[i] = huh.NewOption(k, k)
	}

	var selectedKey string
	keyForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Which secret?").
				Options(keyOpts...).
				Value(&selectedKey),
		),
	)
	if err := keyForm.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return nil
		}
		return err
	}

	value, err := store.GetEntry(&vault, selectedProfile, selectedKey, ident)
	if err != nil {
		iPrintError(fmt.Sprintf("cannot retrieve %s: %v", selectedKey, err))
		return nil
	}
	defer crypto.ZeroBytes(value)

	maskedTitle := fmt.Sprintf("%s  %s",
		iHeaderStyle.Render(selectedKey),
		iSubStyle.Render("••••••••"),
	)
	var action string
	actionForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(maskedTitle).
				Options(
					huh.NewOption("Copy to clipboard", "copy"),
					huh.NewOption("Reveal value  (visible for 5s)", "reveal"),
					huh.NewOption("Back", "back"),
				).
				Value(&action),
		),
	)
	if err := actionForm.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return nil
		}
		return err
	}

	switch action {
	case "copy":
		valueCopy := make([]byte, len(value))
		copy(valueCopy, value)
		if err := copyToClipboard(valueCopy); err != nil {
			iPrintWarn("clipboard not available")
		} else {
			iPrintSuccess(fmt.Sprintf("%s copied to clipboard", ui.Highlight(selectedKey)))
		}
	case "reveal":
		revealBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("214")).
			Padding(0, 2).
			Render(fmt.Sprintf("%s  %s", iHeaderStyle.Render(selectedKey), string(value)))
		fmt.Printf("\n  %s\n\n", revealBox)
		fmt.Printf("  %s\n\n", iSubStyle.Render("masking in 5 seconds…"))
		time.Sleep(5 * time.Second)
		// overwrite the revealed lines with blank space
		fmt.Print("\033[4A\033[0J")
	}
	return nil
}

// RunInteractiveDelete prompts for profile and key then confirms deletion.
func RunInteractiveDelete(_ *cobra.Command) error {
	if !iVaultExists() {
		iPrintError("no vault found in current directory — run dotlock init first")
		return nil
	}

	cfg := config.LoadConfig()
	rec, ident, err := crypto.LoadDefaultRecipientAndIdentity()
	if err != nil {
		iPrintError(fmt.Sprintf("cannot load crypto keys: %v", err))
		return nil
	}

	vault, err := store.LoadVault(filepath.Join(".", cfg.VaultFilename), ident)
	if err != nil {
		iPrintError(fmt.Sprintf("cannot load vault: %v", err))
		return nil
	}

	profiles := iProfileOptions(vault)
	selectedProfile := cfg.DefaultProfile
	profileForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Which profile?").
				Options(profiles...).
				Value(&selectedProfile),
		),
	)
	if err := profileForm.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return nil
		}
		return err
	}

	keys, err := store.ListEntries(&vault, selectedProfile)
	if err != nil {
		iPrintError(fmt.Sprintf("profile %q not found", selectedProfile))
		return nil
	}
	if len(keys) == 0 {
		iPrintWarn(fmt.Sprintf("no secrets in profile %q", selectedProfile))
		return nil
	}
	sort.Strings(keys)

	keyOpts := make([]huh.Option[string], len(keys))
	for i, k := range keys {
		keyOpts[i] = huh.NewOption(k, k)
	}

	var selectedKey string
	keyForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Which secret to delete?").
				Options(keyOpts...).
				Value(&selectedKey),
		),
	)
	if err := keyForm.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return nil
		}
		return err
	}

	var confirm bool
	dangerTitle := iDangerStyle.Render(fmt.Sprintf("Delete %s from %s? This cannot be undone.", selectedKey, selectedProfile))
	confirmForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(dangerTitle).
				Affirmative("Yes, delete").
				Negative("No, cancel").
				Value(&confirm),
		),
	)
	if err := confirmForm.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return nil
		}
		return err
	}
	if !confirm {
		return nil
	}

	if err := store.DeleteEntry(&vault, selectedProfile, selectedKey); err != nil {
		iPrintError(fmt.Sprintf("cannot delete: %v", err))
		return nil
	}

	data, err := store.MarshalAndEncryptVault(&vault, rec)
	if err != nil {
		iPrintError(fmt.Sprintf("cannot encrypt vault: %v", err))
		return nil
	}
	if err := store.AtomicWrite(filepath.Join(".", cfg.VaultFilename), data, 0600); err != nil {
		iPrintError(fmt.Sprintf("cannot write vault: %v", err))
		return nil
	}

	iPrintSuccess(fmt.Sprintf("%s deleted from %s", ui.Highlight(selectedKey), ui.Highlight(selectedProfile)))
	return nil
}

// RunInteractiveDiff prompts for two profiles and shows a coloured diff.
func RunInteractiveDiff(_ *cobra.Command) error {
	if !iVaultExists() {
		iPrintError("no vault found in current directory — run dotlock init first")
		return nil
	}

	cfg := config.LoadConfig()
	_, ident, err := crypto.LoadDefaultRecipientAndIdentity()
	if err != nil {
		iPrintError(fmt.Sprintf("cannot load crypto keys: %v", err))
		return nil
	}

	vault, err := store.LoadVault(filepath.Join(".", cfg.VaultFilename), ident)
	if err != nil {
		iPrintError(fmt.Sprintf("cannot load vault: %v", err))
		return nil
	}

	names := iProfileNames(vault)
	if len(names) < 2 {
		iPrintError("need at least two profiles to diff")
		return nil
	}

	opts := iProfileOptions(vault)
	var profileA string
	formA := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Compare which profile (A)?").
				Options(opts...).
				Value(&profileA),
		),
	)
	if err := formA.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return nil
		}
		return err
	}

	bOpts := make([]huh.Option[string], 0, len(names)-1)
	for _, n := range names {
		if n != profileA {
			bOpts = append(bOpts, huh.NewOption(n, n))
		}
	}
	var profileB string
	formB := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Against which profile (B)?").
				Options(bOpts...).
				Value(&profileB),
		),
	)
	if err := formB.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return nil
		}
		return err
	}

	da, err := store.ProfileMap(&vault, profileA, ident)
	if err != nil {
		iPrintError(fmt.Sprintf("profile %q not found", profileA))
		return nil
	}
	db, err := store.ProfileMap(&vault, profileB, ident)
	if err != nil {
		iPrintError(fmt.Sprintf("profile %q not found", profileB))
		return nil
	}

	iDiffBanner(profileA, profileB)

	lines := diff.Maps(da, db)
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
		sym := strings.TrimSpace(line[:1])
		key := strings.TrimSpace(line[2:])
		switch sym {
		case "+":
			fmt.Printf("  %s  %s\n", iGreenStyle.Render("+"), key)
		case "-":
			fmt.Printf("  %s  %s\n", iRedStyle.Render("-"), key)
		case "~":
			fmt.Printf("  %s  %s\n", iYellowStyle.Render("~"), key)
		default:
			fmt.Printf("  %s  %s\n", iGrayStyle.Render("="), iGrayStyle.Render(key))
		}
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
}

// RunInteractiveExport collects format, profile, and output destination then exports.
func RunInteractiveExport(_ *cobra.Command) error {
	if !iVaultExists() {
		iPrintError("no vault found in current directory — run dotlock init first")
		return nil
	}

	cfg := config.LoadConfig()
	_, ident, err := crypto.LoadDefaultRecipientAndIdentity()
	if err != nil {
		iPrintError(fmt.Sprintf("cannot load crypto keys: %v", err))
		return nil
	}

	vault, err := store.LoadVault(filepath.Join(".", cfg.VaultFilename), ident)
	if err != nil {
		iPrintError(fmt.Sprintf("cannot load vault: %v", err))
		return nil
	}

	profiles := iProfileOptions(vault)
	selectedProfile := cfg.DefaultProfile
	var format string
	var outputDest string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Export format?").
				Options(
					huh.NewOption(`Shell  (export KEY="value")`, "shell"),
					huh.NewOption("Docker (.env file format)", "docker"),
					huh.NewOption("GitHub Actions (gh secret set)", "github"),
				).
				Value(&format),
			huh.NewSelect[string]().
				Title("Which profile?").
				Options(profiles...).
				Value(&selectedProfile),
			huh.NewSelect[string]().
				Title("Output to?").
				Options(
					huh.NewOption("Print to terminal", "terminal"),
					huh.NewOption("Write to file", "file"),
				).
				Value(&outputDest),
		),
	)
	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return nil
		}
		return err
	}

	filePath := ""
	if outputDest == "file" {
		filePath = ".env"
		fileForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("File path").
					Value(&filePath),
			),
		)
		if err := fileForm.Run(); err != nil {
			if errors.Is(err, huh.ErrUserAborted) {
				return nil
			}
			return err
		}

		if _, statErr := os.Stat(filePath); statErr == nil {
			var overwrite bool
			overwriteForm := huh.NewForm(
				huh.NewGroup(
					huh.NewConfirm().
						Title(fmt.Sprintf("%s already exists. Overwrite?", filePath)).
						Affirmative("Yes").
						Negative("No").
						Value(&overwrite),
				),
			)
			if err := overwriteForm.Run(); err != nil {
				if errors.Is(err, huh.ErrUserAborted) {
					return nil
				}
				return err
			}
			if !overwrite {
				return nil
			}
		}
	}

	profileMap, err := store.ProfileMap(&vault, selectedProfile, ident)
	if err != nil {
		iPrintError(fmt.Sprintf("profile %q not found", selectedProfile))
		return nil
	}
	defer func() {
		for _, v := range profileMap {
			crypto.ZeroBytes(v)
		}
	}()

	var output string
	switch format {
	case "shell":
		output = model.RenderShell(profileMap)
	case "docker":
		output = model.RenderDocker(profileMap)
	case "github":
		output = model.RenderGitHubActions(profileMap)
	}

	if outputDest == "terminal" {
		fmt.Print(output)
		return nil
	}

	if err := os.WriteFile(filePath, []byte(output), 0600); err != nil {
		iPrintError(fmt.Sprintf("cannot write file: %v", err))
		return nil
	}
	iPrintSuccess(fmt.Sprintf("exported to %s", ui.Highlight(filePath)))
	return nil
}

// RunInteractiveProfile presents a profile management menu.
func RunInteractiveProfile(_ *cobra.Command) error {
	if !iVaultExists() {
		iPrintError("no vault found in current directory — run dotlock init first")
		return nil
	}

	var action string
	actionForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Profile action?").
				Options(
					huh.NewOption("Switch active profile", "switch"),
					huh.NewOption("Create new profile", "create"),
					huh.NewOption("Delete a profile", "delete"),
					huh.NewOption("List all profiles", "list"),
				).
				Value(&action),
		),
	)
	if err := actionForm.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return nil
		}
		return err
	}

	cfg := config.LoadConfig()
	rec, ident, err := crypto.LoadDefaultRecipientAndIdentity()
	if err != nil {
		iPrintError(fmt.Sprintf("cannot load crypto keys: %v", err))
		return nil
	}

	vault, err := store.LoadVault(filepath.Join(".", cfg.VaultFilename), ident)
	if err != nil {
		iPrintError(fmt.Sprintf("cannot load vault: %v", err))
		return nil
	}

	switch action {
	case "list":
		names := iProfileNames(vault)
		fmt.Println()
		fmt.Printf("  %s  %s  %s  %s\n",
			ui.Dim(fmt.Sprintf("%d profiles", len(names))),
			ui.Dim("·"), ui.Dim("active:"),
			ui.Highlight(cfg.DefaultProfile),
		)
		fmt.Println()
		for _, n := range names {
			ui.ProfileItem(n, n == cfg.DefaultProfile)
		}
		fmt.Println()

	case "switch":
		opts := iProfileOptions(vault)
		selected := cfg.DefaultProfile
		f := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Switch to which profile?").
					Options(opts...).
					Value(&selected),
			),
		)
		if err := f.Run(); err != nil {
			if errors.Is(err, huh.ErrUserAborted) {
				return nil
			}
			return err
		}
		cfg.DefaultProfile = selected
		if err := config.SaveConfig(cfg); err != nil {
			iPrintError(fmt.Sprintf("cannot save config: %v", err))
			return nil
		}
		iPrintSuccess(fmt.Sprintf("switched to %s", ui.Highlight(selected)))

	case "create":
		var newName string
		f := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("New profile name").
					Validate(func(s string) error {
						if !iProfileRegexp.MatchString(s) {
							return fmt.Errorf("must be lowercase letters, numbers, and hyphens only")
						}
						return nil
					}).
					Value(&newName),
			),
		)
		if err := f.Run(); err != nil {
			if errors.Is(err, huh.ErrUserAborted) {
				return nil
			}
			return err
		}
		if err := store.CreateProfile(&vault, newName); err != nil {
			iPrintError(fmt.Sprintf("cannot create profile: %v", err))
			return nil
		}
		data, err := store.MarshalAndEncryptVault(&vault, rec)
		if err != nil {
			iPrintError(fmt.Sprintf("cannot encrypt vault: %v", err))
			return nil
		}
		if err := store.AtomicWrite(filepath.Join(".", cfg.VaultFilename), data, 0600); err != nil {
			iPrintError(fmt.Sprintf("cannot write vault: %v", err))
			return nil
		}
		iPrintSuccess(fmt.Sprintf("profile %s created", ui.Highlight(newName)))

	case "delete":
		opts := iProfileOptions(vault)
		var selected string
		f := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Delete which profile?").
					Options(opts...).
					Value(&selected),
			),
		)
		if err := f.Run(); err != nil {
			if errors.Is(err, huh.ErrUserAborted) {
				return nil
			}
			return err
		}
		var confirm bool
		dangerTitle := iDangerStyle.Render(fmt.Sprintf(
			"Delete profile %s and all its secrets? This cannot be undone.", selected,
		))
		confirmForm := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title(dangerTitle).
					Affirmative("Yes, delete").
					Negative("No, cancel").
					Value(&confirm),
			),
		)
		if err := confirmForm.Run(); err != nil {
			if errors.Is(err, huh.ErrUserAborted) {
				return nil
			}
			return err
		}
		if !confirm {
			return nil
		}
		if err := store.DeleteProfile(&vault, selected); err != nil {
			iPrintError(fmt.Sprintf("cannot delete profile: %v", err))
			return nil
		}
		data, err := store.MarshalAndEncryptVault(&vault, rec)
		if err != nil {
			iPrintError(fmt.Sprintf("cannot encrypt vault: %v", err))
			return nil
		}
		if err := store.AtomicWrite(filepath.Join(".", cfg.VaultFilename), data, 0600); err != nil {
			iPrintError(fmt.Sprintf("cannot write vault: %v", err))
			return nil
		}
		iPrintSuccess(fmt.Sprintf("profile %s deleted", ui.Highlight(selected)))
	}
	return nil
}

// RunInteractiveInit shows a welcome box, collects vault options, then creates the vault.
func RunInteractiveInit(_ *cobra.Command) error {
	cfg := config.LoadConfig()
	vaultPath := filepath.Join(".", cfg.VaultFilename)
	if _, err := os.Stat(vaultPath); err == nil {
		iPrintWarn("vault already exists in this directory")
		return nil
	}

	cwd, _ := os.Getwd()
	defaultName := filepath.Base(cwd)

	welcomeBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(1, 2).
		Render("Welcome to dotlock\nLet's set up your vault")
	fmt.Println()
	fmt.Println(welcomeBox)
	fmt.Println()

	vaultName := ""
	var profileChoice string
	var customProfiles string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Vault name").
				Placeholder(defaultName).
				Value(&vaultName),
			huh.NewSelect[string]().
				Title("Create initial profiles?").
				Options(
					huh.NewOption("dev only", "dev"),
					huh.NewOption("dev + staging + prod", "devstaging"),
					huh.NewOption("custom", "custom"),
				).
				Value(&profileChoice),
		),
	)
	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return nil
		}
		return err
	}

	if profileChoice == "custom" {
		customForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Profile names (comma-separated, e.g. dev,staging,prod)").
					Value(&customProfiles),
			),
		)
		if err := customForm.Run(); err != nil {
			if errors.Is(err, huh.ErrUserAborted) {
				return nil
			}
			return err
		}
	}

	var confirmCreate bool
	confirmForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Create vault?").
				Affirmative("Yes").
				Negative("No").
				Value(&confirmCreate),
		),
	)
	if err := confirmForm.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return nil
		}
		return err
	}
	if !confirmCreate {
		return nil
	}

	recipient, _, err := crypto.LoadDefaultRecipientAndIdentity()
	if err != nil {
		iPrintError(fmt.Sprintf("cannot load or create identity: %v", err))
		return nil
	}

	if vaultName == "" {
		vaultName = defaultName
	}

	vault := store.NewEmptyVault()
	vault.ID = uuid.New()
	vault.Name = vaultName
	vault.CreatedAt = time.Now().UTC()
	vault.UpdatedAt = vault.CreatedAt

	switch profileChoice {
	case "dev":
		vault.Profiles = map[string]store.Profile{
			"dev": {Entries: map[string]store.Entry{}},
		}
	case "devstaging":
		vault.Profiles = map[string]store.Profile{
			"dev":     {Entries: map[string]store.Entry{}},
			"staging": {Entries: map[string]store.Entry{}},
			"prod":    {Entries: map[string]store.Entry{}},
		}
	case "custom":
		vault.Profiles = map[string]store.Profile{}
		for _, p := range strings.Split(customProfiles, ",") {
			p = strings.TrimSpace(p)
			if p != "" {
				vault.Profiles[p] = store.Profile{Entries: map[string]store.Entry{}}
			}
		}
		if len(vault.Profiles) == 0 {
			vault.Profiles["dev"] = store.Profile{Entries: map[string]store.Entry{}}
		}
	}

	data, err := store.MarshalAndEncryptVault(&vault, recipient)
	if err != nil {
		iPrintError(fmt.Sprintf("cannot encrypt vault: %v", err))
		return nil
	}
	if err := store.AtomicWrite(vaultPath, data, 0600); err != nil {
		iPrintError(fmt.Sprintf("cannot write vault: %v", err))
		return nil
	}

	fmt.Println()
	iPrintSuccess("Vault created")
	fmt.Println()
	fmt.Printf("  %s\n", ui.Dim("Next steps:"))
	fmt.Printf("  %-28s  %s\n", ui.Highlight("dotlock set DATABASE_URL"), ui.Dim("add your first secret"))
	fmt.Printf("  %-28s  %s\n", ui.Highlight("dotlock ui"), ui.Dim("open the visual interface"))
	fmt.Println()
	return nil
}

// copyToClipboard writes value to the OS clipboard and zeros the slice after.
func copyToClipboard(value []byte) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbcopy")
	case "windows":
		cmd = exec.Command("clip")
	default:
		cmd = exec.Command("xclip", "-selection", "clipboard")
	}
	pipe, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	if _, err := pipe.Write(value); err != nil {
		pipe.Close()
		return err
	}
	pipe.Close()
	runErr := cmd.Wait()
	for i := range value {
		value[i] = 0
	}
	return runErr
}
