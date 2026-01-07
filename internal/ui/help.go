package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/rk/tgcp/internal/styles"
)

// HelpView renders the help screen overlay
func HelpView(width, height int) string {
	doc := strings.Builder{}

	// Header
	doc.WriteString(styles.TitleStyle.Render("TGCP Help & Keybindings"))
	doc.WriteString("\n\n")

	// Sections
	globalKeys := [][]string{
		{":", "Open Command Palette"},
		{"?", "Toggle Help Screen"},
		{"q", "Back / Quit"},
		{"ctrl+c", "Force Quit"},
		{"Tab", "Toggle Sidebar"},
	}

	navigationKeys := [][]string{
		{"↑/↓ or j/k", "Navigate List"},
		{"Enter", "Select / View Details"},
		{"/", "Filter List"},
		{"n/p", "Next/Previous Page"},
	}

	// Helper to render key table
	renderTable := func(title string, keys [][]string) string {
		s := strings.Builder{}
		s.WriteString(styles.HeaderStyle.Render(title))
		s.WriteString("\n")

		keyStyle := lipgloss.NewStyle().Foreground(styles.ColorPrimary).Bold(true).Width(15)
		descStyle := lipgloss.NewStyle().Foreground(styles.ColorText)

		for _, row := range keys {
			s.WriteString(fmt.Sprintf("%s %s\n", keyStyle.Render(row[0]), descStyle.Render(row[1])))
		}
		return s.String()
	}

	// Columns
	col1 := renderTable("Global Navigation", globalKeys)
	col2 := renderTable("List View", navigationKeys)

	// Layout
	content := lipgloss.JoinHorizontal(lipgloss.Top, col1, "    ", col2)

	// Box it
	dialog := styles.BoxStyle.Copy().
		Width(60).
		BorderForeground(styles.ColorHighlight).
		Render(content)

	// Center in screen
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, dialog)
}
