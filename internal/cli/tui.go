package cli

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/ahmadraza100/dotlock/internal/tui"
)

var uiCmd = &cobra.Command{
	Use:   "ui",
	Short: "launch the visual vault manager",
	RunE: func(_ *cobra.Command, _ []string) error {
		p := tea.NewProgram(tui.NewApp(), tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("cannot start tui: %w", err)
		}
		return nil
	},
}
