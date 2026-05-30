package tui

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ahmadraza100/dotlock/internal/config"
	"github.com/ahmadraza100/dotlock/internal/crypto"
	"github.com/ahmadraza100/dotlock/internal/model"
	"github.com/ahmadraza100/dotlock/internal/store"
	views "github.com/ahmadraza100/dotlock/internal/tui/views"
	"github.com/ahmadraza100/dotlock/pkg/diff"
)

const (
	paneProfiles = iota
	paneSecrets
)

// ASCII logo
const logoASCII = `____  ____ _____ _     ____  ____ _  __
/  _ \/  _ Y__ __Y \   /  _ \/   _Y |/ /
| | \|| / \| / \ | |   | / \||  / |   / 
| |_/|| \_/| | | | |_/\| \_/||  \_|   \ 
\____/\____/ \_/ \____/\____/\____|_|\_\
                                        `

type secretReveal struct {
	key    string
	value  []byte
	expire time.Time
}

// App is the root BubbleTea model.
type App struct {
	ctx context.Context

	profiles list.Model
	secrets  list.Model

	currentVault     store.Vault
	currentVaultPath string
	currentProfile   string

	vaultFilename string

	activePane    int
	width, height int

	revealed *secretReveal

	// mode drives the input prompt
	// "" | "newKey" | "newValue" | "editValue" | "confirmDelete" |
	// "addProfile" | "confirmDeleteProfile" | "importPath" | "exportPath"
	mode        string
	ti          textinput.Model
	modalKey    string
	modalPrompt string
	status      string
	statusErr   bool
	showHelp    bool

	// diff overlay
	showDiff   bool
	diffTarget string
}

// ---------------------------------------------------------------------------
// List item types
// ---------------------------------------------------------------------------

type profileItem struct {
	name  string
	count int
}

func (p profileItem) Title() string { return p.name }
func (p profileItem) Description() string {
	switch p.count {
	case 0:
		return "no secrets"
	case 1:
		return "1 secret"
	default:
		return fmt.Sprintf("%d secrets", p.count)
	}
}
func (p profileItem) FilterValue() string { return p.name }

type secretItem struct{ key string }

func (s secretItem) Title() string       { return s.key }
func (s secretItem) Description() string { return "  encrypted" }
func (s secretItem) FilterValue() string { return s.key }

// ---------------------------------------------------------------------------
// Init
// ---------------------------------------------------------------------------

func NewApp() *App {
	invisible := lipgloss.NewStyle()

	mkList := func(w, h int, showDesc bool) list.Model {
		d := list.NewDefaultDelegate()
		d.ShowDescription = showDesc
		l := list.New([]list.Item{}, d, w, h)
		l.Title = ""
		l.Styles.Title = invisible
		l.SetShowStatusBar(false)
		l.SetFilteringEnabled(true)
		l.SetShowHelp(false)
		return l
	}

	ti := textinput.New()
	ti.CharLimit = 512

	cfg := config.LoadConfig()

	a := &App{
		ctx:           context.Background(),
		profiles:      mkList(28, 20, true),
		secrets:       mkList(58, 20, true),
		activePane:    paneProfiles,
		ti:            ti,
		vaultFilename: cfg.VaultFilename,
	}
	a.loadCurrentVault()
	return a
}

func (a *App) Init() tea.Cmd { return nil }

// ---------------------------------------------------------------------------
// Update
// ---------------------------------------------------------------------------

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		return a, nil

	case tea.KeyMsg:
		// modal captures all keys
		if a.mode != "" {
			return a.updateModal(msg)
		}

		// diff overlay captures navigation and esc
		if a.showDiff {
			switch msg.String() {
			case "esc", "f":
				a.showDiff = false
			case "right", "l":
				a.cycleDiffTarget(+1)
			case "left", "h":
				a.cycleDiffTarget(-1)
			}
			return a, nil
		}

		a.status = ""
		a.statusErr = false

		switch msg.String() {
		// navigation
		case "tab":
			a.activePane = (a.activePane + 1) % 2
			return a, nil
		case "shift+tab":
			a.activePane = (a.activePane + 1) % 2
			return a, nil
		case "q", "ctrl+c":
			return a, tea.Quit
		case "?":
			a.showHelp = !a.showHelp
			return a, nil

		// --- secrets ---
		case "n":
			if a.currentVaultPath != "" && a.currentProfile != "" {
				a.mode = "newKey"
				a.modalPrompt = "new secret key"
				a.ti.SetValue("")
				a.ti.Placeholder = "DATABASE_URL"
				a.activePane = paneSecrets
				return a, a.ti.Focus()
			}
			a.setErr("select a vault and profile first")
			return a, nil

		case "e":
			if a.activePane == paneSecrets {
				if sel := a.secrets.SelectedItem(); sel != nil {
					a.modalKey = sel.FilterValue()
					a.mode = "editValue"
					a.modalPrompt = "edit value for  " + a.modalKey
					a.ti.SetValue("")
					a.ti.Placeholder = "new value"
					return a, a.ti.Focus()
				}
			}

		case "v":
			if a.activePane == paneSecrets {
				if sel := a.secrets.SelectedItem(); sel != nil {
					a.revealSecret(sel.FilterValue())
				}
			}

		case "d":
			switch a.activePane {
			case paneSecrets:
				if sel := a.secrets.SelectedItem(); sel != nil {
					a.modalKey = sel.FilterValue()
					a.modalPrompt = "delete secret  " + a.modalKey + "?"
					a.mode = "confirmDelete"
				}
			}

		// --- profiles ---
		case "a":
			switch a.activePane {
			case paneProfiles:
				if a.currentVaultPath != "" {
					a.mode = "addProfile"
					a.modalPrompt = "new profile name"
					a.ti.SetValue("")
					a.ti.Placeholder = "staging"
					return a, a.ti.Focus()
				}
			case paneSecrets:
				if a.currentVaultPath != "" && a.currentProfile != "" {
					a.mode = "newKey"
					a.modalPrompt = "new secret key"
					a.ti.SetValue("")
					a.ti.Placeholder = "DATABASE_URL"
					return a, a.ti.Focus()
				}
			}

		case "D": // shift+D = delete profile
			if a.activePane == paneProfiles {
				if sel := a.profiles.SelectedItem(); sel != nil {
					a.modalKey = sel.FilterValue()
					a.modalPrompt = "delete profile  " + a.modalKey + "  and all its secrets?"
					a.mode = "confirmDeleteProfile"
				}
			}

		// --- vault ---
		case "i":
			if a.currentVaultPath != "" && a.currentProfile != "" {
				a.mode = "importPath"
				a.modalPrompt = "import .env file path"
				a.ti.SetValue(".env")
				a.ti.Placeholder = ".env"
				return a, a.ti.Focus()
			}
			a.setErr("select a vault and profile first")

		case "x":
			if a.currentVaultPath != "" && a.currentProfile != "" {
				a.mode = "exportPath"
				a.modalPrompt = "export to file path"
				a.ti.SetValue(a.currentProfile + ".env")
				a.ti.Placeholder = a.currentProfile + ".env"
				return a, a.ti.Focus()
			}
			a.setErr("select a vault and profile first")

		case "f":
			if a.currentVaultPath == "" || a.currentProfile == "" {
				a.setErr("select a vault and profile first")
				return a, nil
			}
			names := a.sortedProfileNames()
			if len(names) < 2 {
				a.setErr("need at least 2 profiles to diff")
				return a, nil
			}
			// default diffTarget: the next profile after currentProfile
			a.diffTarget = a.nextProfile(names, a.currentProfile, +1)
			a.showDiff = true
			return a, nil
		}
	}

	// forward to active list
	var cmd tea.Cmd
	switch a.activePane {
	case paneProfiles:
		prev := a.currentProfile
		a.profiles, cmd = a.profiles.Update(msg)
		if s := a.profiles.SelectedItem(); s != nil && s.FilterValue() != prev {
			a.loadSecretsForProfile(s.FilterValue())
		}
	case paneSecrets:
		a.secrets, cmd = a.secrets.Update(msg)
	}

	if a.revealed != nil {
		return a, tea.Batch(cmd, tea.Tick(time.Until(a.revealed.expire), func(t time.Time) tea.Msg {
			return t
		}))
	}
	return a, cmd
}

func (a *App) updateModal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		return a, a.handleModalEnter()
	case "esc":
		a.mode = ""
		a.ti.SetValue("")
		return a, nil
	default:
		var cmd tea.Cmd
		a.ti, cmd = a.ti.Update(msg)
		return a, cmd
	}
}

func (a *App) handleModalEnter() tea.Cmd {
	defer func() { a.ti.SetValue("") }()

	switch a.mode {
	// ---- two-step secret creation ----
	case "newKey":
		key := a.ti.Value()
		if !isValidKey(key) {
			a.setErr("invalid key — use A-Z, digits, underscores only")
			a.mode = ""
			return nil
		}
		a.modalKey = key
		a.mode = "newValue"
		a.modalPrompt = "value for  " + key
		a.ti.Placeholder = "secret value"
		return a.ti.Focus()

	case "newValue":
		val := []byte(a.ti.Value())
		if err := a.setEntry(a.modalKey, val); err != nil {
			a.setErr(err.Error())
		} else {
			a.status = fmt.Sprintf("saved  %s", a.modalKey)
			a.reloadProfiles()
			a.loadSecretsForProfile(a.currentProfile)
		}
		crypto.ZeroBytes(val)
		a.mode = ""

	// ---- edit secret ----
	case "editValue":
		val := []byte(a.ti.Value())
		if err := a.setEntry(a.modalKey, val); err != nil {
			a.setErr(err.Error())
		} else {
			a.status = fmt.Sprintf("updated  %s", a.modalKey)
			a.reloadProfiles()
			a.loadSecretsForProfile(a.currentProfile)
		}
		crypto.ZeroBytes(val)
		a.mode = ""

	// ---- delete secret ----
	case "confirmDelete":
		if err := a.deleteEntry(a.modalKey); err != nil {
			a.setErr(err.Error())
		} else {
			a.status = fmt.Sprintf("deleted  %s", a.modalKey)
			a.reloadProfiles()
			a.loadSecretsForProfile(a.currentProfile)
		}
		a.mode = ""

	// ---- add profile ----
	case "addProfile":
		name := strings.TrimSpace(a.ti.Value())
		if !profileRegexp.MatchString(name) {
			a.setErr("use lowercase letters, numbers, hyphens only")
			a.mode = ""
			return nil
		}
		if err := a.createProfile(name); err != nil {
			a.setErr(err.Error())
		} else {
			a.status = fmt.Sprintf("profile %s created", name)
			a.reloadProfiles()
		}
		a.mode = ""

	// ---- delete profile ----
	case "confirmDeleteProfile":
		if err := a.deleteProfile(a.modalKey); err != nil {
			a.setErr(err.Error())
		} else {
			a.status = fmt.Sprintf("profile %s deleted", a.modalKey)
			a.reloadProfiles()
			a.currentProfile = ""
			a.secrets.SetItems(nil)
		}
		a.mode = ""

	// ---- import .env ----
	case "importPath":
		path := strings.TrimSpace(a.ti.Value())
		if err := a.importEnvFile(path); err != nil {
			a.setErr(err.Error())
		}
		a.mode = ""

	// ---- export .env ----
	case "exportPath":
		path := strings.TrimSpace(a.ti.Value())
		if err := a.exportToFile(path); err != nil {
			a.setErr(err.Error())
		} else {
			a.status = fmt.Sprintf("exported to %s", path)
		}
		a.mode = ""
	}
	return nil
}

// ---------------------------------------------------------------------------
// View
// ---------------------------------------------------------------------------

func (a *App) layout() (w, h, pW1, pW2, boxInner int) {
	w = a.width
	if w < 60 {
		w = 60
	}
	// Header/controls boxes: border(2) + padding left+right(4) = 6 overhead
	boxInner = w - 6

	// 2 panels × 4 overhead each = 8 total
	usable := w - 8
	if usable < 30 {
		usable = 30
	}
	pW1 = usable / 3   // profiles: ~1/3
	pW2 = usable - pW1 // secrets:  ~2/3

	headerLines := 10
	if w < 72 {
		headerLines = 5
	}
	overhead := headerLines + 7 + 1 + 4
	h = a.height - overhead
	if h < 4 {
		h = 4
	}
	return
}

func (a *App) View() string {
	w, h, pW1, pW2, boxInner := a.layout()

	a.profiles.SetSize(pW1, h)
	a.secrets.SetSize(pW2, h)

	header := a.headerView(w, boxInner)

	pane := func(idx, pw int, label string, listView string) string {
		lbl := paneLabelInactive.Render(label)
		if idx == a.activePane {
			lbl = paneLabelActive.Render(label)
		}
		inner := lbl + "\n" + listView
		s := inactivePanelStyle.Width(pw)
		if idx == a.activePane {
			s = activePanelStyle.Width(pw)
		}
		return s.Render(inner)
	}

	panels := lipgloss.JoinHorizontal(lipgloss.Top,
		pane(paneProfiles, pW1, "  profiles", a.profiles.View()),
		pane(paneSecrets, pW2, "  secrets", a.secrets.View()),
	)

	controls := a.controlsView(boxInner)

	statusLine := ""
	if a.revealed != nil {
		statusLine = "\n  " + revealedKeyStyle.Render(a.revealed.key) +
			"  " + revealedValStyle.Render(string(a.revealed.value)) +
			taglineStyle.Render("  (clears in 30s)")
	} else if a.status != "" {
		if a.statusErr {
			statusLine = "\n  " + statusErrStyle.Render("✗  "+a.status)
		} else {
			statusLine = "\n  " + statusKeyStyle.Render("✓  "+a.status)
		}
	}

	body := header + panels + "\n" + controls + statusLine

	if a.showHelp {
		title := helpSectionStyle.Render("keyboard shortcuts")
		return body + "\n" + helpStyle.Render(title+"\n\n"+views.HelpView())
	}
	if a.showDiff {
		return body + "\n" + a.diffView()
	}
	if a.mode != "" {
		return body + "\n" + modalStyle.Render(a.modalView())
	}
	return body
}

func (a *App) headerView(w, boxInner int) string {
	var logoLine string
	if w >= 72 {
		logoLine = logoStyle.Render(logoASCII) + "\n"
	}

	tag := taglineStyle.Render("encrypted .env vault manager")
	ctx := ""
	if a.currentVaultPath != "" {
		vaultLabel := a.currentVault.Name
		if vaultLabel == "" {
			vaultLabel = filepath.Base(a.currentVaultPath)
		}
		ctx = "  ·  " + contextStyle.Render(vaultLabel)
		if a.currentProfile != "" {
			ctx += taglineStyle.Render(" / ") + contextStyle.Render(a.currentProfile)
		}
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(brandColor).
		Padding(0, 2).
		Width(boxInner).
		Render(logoLine+tag+ctx) + "\n"
}

func (a *App) controlsView(boxInner int) string {
	k := func(key, desc string) string {
		return controlsKeyStyle.Render("["+key+"]") + " " + controlsDescStyle.Render(desc)
	}
	sep := controlsDescStyle.Render("  ·  ")

	rows := strings.Join([]string{
		strings.Join([]string{k("n", "new secret"), k("e", "edit"), k("v", "reveal"), k("d", "del secret")}, sep),
		strings.Join([]string{k("a", "add profile"), k("D", "del profile"), k("i", "import"), k("x", "export"), k("f", "diff")}, sep),
		strings.Join([]string{k("tab", "switch pane"), k("/", "search"), k("?", "help"), k("q", "quit")}, sep),
	}, "\n")

	return controlsBoxStyle.Width(boxInner).Render(
		controlsSectionStyle.Render("controls")+"\n"+rows,
	) + "\n"
}

func (a *App) modalView() string {
	prompt := brandNameStyle.Render(a.modalPrompt)
	switch a.mode {
	case "confirmDelete", "confirmDeleteProfile":
		return prompt + "\n\n" +
			taglineStyle.Render("↵ confirm   ·   esc cancel")
	case "newValue", "editValue":
		disp := a.ti.View()
		if a.ti.Value() != "" {
			disp = taglineStyle.Render("  ••••••••")
		}
		return prompt + "\n\n" + disp + "\n\n" +
			taglineStyle.Render("↵ save   ·   esc cancel")
	default:
		return prompt + "\n\n" + a.ti.View() + "\n\n" +
			taglineStyle.Render("↵ confirm   ·   esc cancel")
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func (a *App) setErr(msg string) {
	a.status = msg
	a.statusErr = true
}

// sortedProfileNames returns all profile names in sorted order.
func (a *App) sortedProfileNames() []string {
	names := make([]string, 0, len(a.currentVault.Profiles))
	for n := range a.currentVault.Profiles {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// nextProfile returns the profile that is `step` positions away from current in
// the sorted names list (wraps around).
func (a *App) nextProfile(names []string, current string, step int) string {
	if len(names) == 0 {
		return ""
	}
	idx := 0
	for i, n := range names {
		if n == current {
			idx = i
			break
		}
	}
	idx = ((idx+step)%len(names) + len(names)) % len(names)
	return names[idx]
}

// cycleDiffTarget shifts diffTarget by step positions in the profile list,
// skipping currentProfile so we never diff a profile against itself.
func (a *App) cycleDiffTarget(step int) {
	names := a.sortedProfileNames()
	if len(names) < 2 {
		return
	}
	// collect candidates (all except currentProfile)
	candidates := make([]string, 0, len(names)-1)
	for _, n := range names {
		if n != a.currentProfile {
			candidates = append(candidates, n)
		}
	}
	idx := 0
	for i, n := range candidates {
		if n == a.diffTarget {
			idx = i
			break
		}
	}
	idx = ((idx+step)%len(candidates) + len(candidates)) % len(candidates)
	a.diffTarget = candidates[idx]
}

// diffView renders the color-coded diff overlay between currentProfile and diffTarget.
func (a *App) diffView() string {
	_, ident, err := crypto.LoadDefaultRecipientAndIdentity()
	if err != nil {
		return diffBoxStyle.Render(diffTitleStyle.Render("diff error") + "\n\n" + err.Error())
	}

	mapA, err := store.ProfileMap(&a.currentVault, a.currentProfile, ident)
	if err != nil {
		return diffBoxStyle.Render(diffTitleStyle.Render("diff error") + "\n\n" + err.Error())
	}
	mapB, err := store.ProfileMap(&a.currentVault, a.diffTarget, ident)
	if err != nil {
		// zero mapA before returning
		for _, v := range mapA {
			crypto.ZeroBytes(v)
		}
		return diffBoxStyle.Render(diffTitleStyle.Render("diff error") + "\n\n" + err.Error())
	}

	lines := diff.Maps(mapA, mapB)

	// zero the decrypted maps immediately after use
	defer func() {
		for _, v := range mapA {
			crypto.ZeroBytes(v)
		}
		for _, v := range mapB {
			crypto.ZeroBytes(v)
		}
	}()

	var sb strings.Builder
	title := diffTitleStyle.Render(fmt.Sprintf("diff  %s  →  %s", a.currentProfile, a.diffTarget))
	hint := taglineStyle.Render("  [←][→] cycle target   esc close")
	sb.WriteString(title + hint + "\n\n")

	if len(lines) == 0 {
		sb.WriteString(taglineStyle.Render("(no keys in either profile)"))
	} else {
		for _, line := range lines {
			if len(line) < 2 {
				sb.WriteString(taglineStyle.Render(line) + "\n")
				continue
			}
			prefix := line[:2]
			rest := line[2:]
			switch prefix {
			case "+ ":
				sb.WriteString(diffAddStyle.Render("+  "+rest) + "\n")
			case "- ":
				sb.WriteString(diffDelStyle.Render("-  "+rest) + "\n")
			case "~ ":
				sb.WriteString(diffChgStyle.Render("~  "+rest) + "\n")
			default:
				sb.WriteString(diffEqStyle.Render("   "+rest) + "\n")
			}
		}
	}

	return diffBoxStyle.Render(sb.String())
}

var keyNameRegexp = regexp.MustCompile(`^[A-Z_][A-Z0-9_]*$`)
var profileRegexp = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

func isValidKey(k string) bool {
	return keyNameRegexp.MatchString(k) && len(k) <= 256
}

func (a *App) reloadProfiles() {
	if a.currentVaultPath == "" {
		return
	}
	names := make([]string, 0, len(a.currentVault.Profiles))
	for n := range a.currentVault.Profiles {
		names = append(names, n)
	}
	sort.Strings(names)
	items := make([]list.Item, 0, len(names))
	for _, n := range names {
		items = append(items, profileItem{name: n, count: len(a.currentVault.Profiles[n].Entries)})
	}
	a.profiles.SetItems(items)
}

func (a *App) persistVault() error {
	rec, _, err := crypto.LoadDefaultRecipientAndIdentity()
	if err != nil {
		return fmt.Errorf("cannot load keys: %w", err)
	}
	data, err := store.MarshalAndEncryptVault(&a.currentVault, rec)
	if err != nil {
		return err
	}
	return store.AtomicWrite(filepath.Join(a.currentVaultPath, a.vaultFilename), data, 0600)
}

// ---------------------------------------------------------------------------
// Vault operations
// ---------------------------------------------------------------------------

func (a *App) setEntry(key string, value []byte) error {
	if len(value) > 65536 {
		return fmt.Errorf("value too long")
	}
	rec, _, err := crypto.LoadDefaultRecipientAndIdentity()
	if err != nil {
		return fmt.Errorf("cannot load keys: %w", err)
	}
	if err := store.SetEntry(&a.currentVault, a.currentProfile, key, value, rec); err != nil {
		return err
	}
	return a.persistVault()
}

func (a *App) deleteEntry(key string) error {
	if err := store.DeleteEntry(&a.currentVault, a.currentProfile, key); err != nil {
		return err
	}
	return a.persistVault()
}

func (a *App) createProfile(name string) error {
	if err := store.CreateProfile(&a.currentVault, name); err != nil {
		return err
	}
	return a.persistVault()
}

func (a *App) deleteProfile(name string) error {
	if err := store.DeleteProfile(&a.currentVault, name); err != nil {
		return err
	}
	return a.persistVault()
}

func (a *App) revealSecret(key string) {
	_, ident, err := crypto.LoadDefaultRecipientAndIdentity()
	if err != nil {
		a.setErr("cannot load keys")
		return
	}
	val, err := store.GetEntry(&a.currentVault, a.currentProfile, key, ident)
	if err != nil {
		a.setErr("cannot decrypt: " + err.Error())
		return
	}
	if a.revealed != nil {
		crypto.ZeroBytes(a.revealed.value)
	}
	a.revealed = &secretReveal{key: key, value: val, expire: time.Now().Add(30 * time.Second)}
}

func (a *App) importEnvFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("cannot open %s: %w", path, err)
	}
	defer f.Close()

	rec, _, err := crypto.LoadDefaultRecipientAndIdentity()
	if err != nil {
		return err
	}

	count := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.IndexByte(line, '=')
		if idx <= 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		if len(val) >= 2 && ((val[0] == '"' && val[len(val)-1] == '"') ||
			(val[0] == '\'' && val[len(val)-1] == '\'')) {
			val = val[1 : len(val)-1]
		}
		if !isValidKey(key) {
			continue
		}
		valB := []byte(val)
		if err := store.SetEntry(&a.currentVault, a.currentProfile, key, valB, rec); err == nil {
			count++
		}
		crypto.ZeroBytes(valB)
	}
	if err := a.persistVault(); err != nil {
		return err
	}
	a.status = fmt.Sprintf("imported %d secrets from %s", count, filepath.Base(path))
	a.loadSecretsForProfile(a.currentProfile)
	a.reloadProfiles()
	return nil
}

func (a *App) exportToFile(path string) error {
	_, ident, err := crypto.LoadDefaultRecipientAndIdentity()
	if err != nil {
		return err
	}
	pm, err := store.ProfileMap(&a.currentVault, a.currentProfile, ident)
	if err != nil {
		return err
	}
	defer func() {
		for _, v := range pm {
			crypto.ZeroBytes(v)
		}
	}()
	var content string
	if strings.HasSuffix(path, ".sh") {
		content = model.RenderShell(pm)
	} else {
		content = model.RenderDocker(pm)
	}
	return os.WriteFile(path, []byte(content), 0600)
}

// ---------------------------------------------------------------------------
// Vault loading
// ---------------------------------------------------------------------------

func (a *App) loadCurrentVault() {
	cwd, _ := os.Getwd()
	abs, _ := filepath.Abs(cwd)
	vaultPath := filepath.Join(abs, a.vaultFilename)

	_, ident, err := crypto.LoadDefaultRecipientAndIdentity()
	if err != nil {
		a.setErr("cannot load keys: " + err.Error())
		return
	}
	v, err := store.LoadVault(vaultPath, ident)
	if err != nil {
		a.setErr("no vault here — run: dotlock init")
		return
	}
	a.currentVault = v
	a.currentVaultPath = abs

	names := make([]string, 0, len(v.Profiles))
	for n := range v.Profiles {
		names = append(names, n)
	}
	sort.Strings(names)

	items := make([]list.Item, 0, len(names))
	for _, n := range names {
		items = append(items, profileItem{name: n, count: len(v.Profiles[n].Entries)})
	}
	a.profiles.SetItems(items)

	if len(names) > 0 {
		a.loadSecretsForProfile(names[0])
	}
}

func (a *App) loadSecretsForProfile(profile string) {
	prof, ok := a.currentVault.Profiles[profile]
	if !ok {
		a.secrets.SetItems(nil)
		return
	}
	a.currentProfile = profile

	keys := make([]string, 0, len(prof.Entries))
	for k := range prof.Entries {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	items := make([]list.Item, 0, len(keys))
	for _, k := range keys {
		items = append(items, secretItem{key: k})
	}
	a.secrets.SetItems(items)
}
