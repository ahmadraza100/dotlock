package tui

import "github.com/charmbracelet/lipgloss"

var (
	brandColor   = lipgloss.Color("63")
	accentColor  = lipgloss.Color("81")
	mutedColor   = lipgloss.Color("240")
	subtleColor  = lipgloss.Color("238")
	successColor = lipgloss.Color("78")
	warnColor    = lipgloss.Color("214")

	// header
	logoStyle = lipgloss.NewStyle().
			Foreground(brandColor).
			Bold(true)

	brandNameStyle = lipgloss.NewStyle().
			Foreground(brandColor).
			Bold(true)

	taglineStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	contextStyle = lipgloss.NewStyle().
			Foreground(accentColor)

	headerStyle = lipgloss.NewStyle().Bold(true).Foreground(brandColor)

	headerBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(brandColor).
			Padding(0, 2)

	// panels
	activePanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(accentColor).
				Padding(0, 1)

	inactivePanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(subtleColor).
				Padding(0, 1)

	paneLabelActive = lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true)

	paneLabelInactive = lipgloss.NewStyle().
				Foreground(mutedColor)

	// controls panel
	controlsBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(subtleColor).
				Padding(0, 2)

	controlsSectionStyle = lipgloss.NewStyle().
				Foreground(mutedColor).
				Bold(true)

	controlsKeyStyle = lipgloss.NewStyle().
				Foreground(accentColor).
				Bold(true)

	controlsDescStyle = lipgloss.NewStyle().
				Foreground(mutedColor)

	// modal
	modalStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(brandColor).
			Padding(1, 3).
			Width(62).
			Align(lipgloss.Left)

	// help
	helpStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(accentColor).
			Foreground(lipgloss.Color("252")).
			Padding(1, 2).
			Width(52)

	helpSectionStyle = lipgloss.NewStyle().
				Foreground(accentColor).
				Bold(true)

	helpKeyStyle = lipgloss.NewStyle().
			Foreground(brandColor).
			Bold(true)

	// status / status inline
	statusKeyStyle = lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true)

	statusErrStyle = lipgloss.NewStyle().
			Foreground(warnColor).
			Bold(true)

	// reveal
	revealedKeyStyle = lipgloss.NewStyle().
				Foreground(accentColor).
				Bold(true)

	revealedValStyle = lipgloss.NewStyle().
				Foreground(successColor).
				Bold(true)

	revealBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(successColor).
			Padding(0, 2)

	// diff overlay
	diffBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(accentColor).
			Foreground(lipgloss.Color("252")).
			Padding(1, 3).
			Width(72)

	diffTitleStyle = lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true)

	diffAddStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("78")) // green

	diffDelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")) // red

	diffChgStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")) // yellow/amber

	diffEqStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")) // dim gray
)
