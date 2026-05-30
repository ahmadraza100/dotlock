package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	green  = lipgloss.NewStyle().Foreground(lipgloss.Color("78")).Bold(true)
	red    = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	yellow = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
	cyan   = lipgloss.NewStyle().Foreground(lipgloss.Color("81")).Bold(true)
	gray   = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	white  = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	bold   = lipgloss.NewStyle().Bold(true)
)

// Success prints a green check with a message.
func Success(msg string) {
	fmt.Printf("  %s  %s\n", green.Render("✓"), msg)
}

// Error prints a red cross with a message.
func Error(msg string) {
	fmt.Printf("  %s  %s\n", red.Render("✗"), msg)
}

// Info prints a neutral info line.
func Info(msg string) {
	fmt.Printf("  %s  %s\n", gray.Render("·"), msg)
}

// Warn prints a yellow warning.
func Warn(msg string) {
	fmt.Printf("  %s  %s\n", yellow.Render("!"), msg)
}

// ProfileHeader prints the context line shown above secret listings.
func ProfileHeader(profile string, count int) {
	fmt.Println()
	fmt.Printf("  %s  %s  %s  %s\n",
		gray.Render("profile"),
		cyan.Render(profile),
		gray.Render("·"),
		gray.Render(fmt.Sprintf("%d secrets", count)),
	)
	fmt.Println()
}

// ProfileHeaderNoCount prints the profile line without a count.
func ProfileHeaderNoCount(profile string) {
	fmt.Println()
	fmt.Printf("  %s  %s\n", gray.Render("profile"), cyan.Render(profile))
	fmt.Println()
}

// SecretRow prints a masked secret row.
func SecretRow(key string) {
	fmt.Printf("  %-32s  %s\n", white.Render(key), gray.Render("••••••••••"))
}

// SecretRowValue prints a revealed secret row.
func SecretRowValue(key, value string) {
	fmt.Printf("  %-32s  %s\n", white.Render(key), value)
}

// Rule prints a thin horizontal separator.
func Rule() {
	fmt.Printf("  %s\n", gray.Render(strings.Repeat("─", 46)))
}

// ProfileItem prints one line in a profiles list.
func ProfileItem(name string, active bool) {
	if active {
		fmt.Printf("  %s  %s  %s\n", green.Render("●"), bold.Render(name), gray.Render("active"))
	} else {
		fmt.Printf("  %s  %s\n", gray.Render("○"), name)
	}
}

// DiffLine prints one colored diff line.
func DiffLine(symbol, key string) {
	switch symbol {
	case "+":
		fmt.Printf("  %s  %s\n", green.Render("+"), white.Render(key))
	case "-":
		fmt.Printf("  %s  %s\n", red.Render("-"), white.Render(key))
	case "~":
		fmt.Printf("  %s  %s\n", yellow.Render("~"), white.Render(key))
	default:
		fmt.Printf("  %s  %s\n", gray.Render("="), gray.Render(key))
	}
}

// KV prints a label → value pair, used in init output and similar.
func KV(label, value string) {
	fmt.Printf("  %-20s  %s\n", gray.Render(label), white.Render(value))
}

// Blank prints an empty line.
func Blank() {
	fmt.Println()
}

// Dim returns text rendered in gray.
func Dim(text string) string {
	return gray.Render(text)
}

// Highlight returns text rendered in cyan.
func Highlight(text string) string {
	return cyan.Render(text)
}
