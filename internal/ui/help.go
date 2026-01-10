package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/yogirk/tgcp/internal/styles"
)

// HelpView renders the help screen overlay
func HelpView(width, height int) string {
	// Define keybindings data
	sections := []struct {
		title string
		items [][]string
	}{
		{
			"Global",
			[][]string{
				{":", "Command Palette"},
				{"?", "Toggle Help"},
				{"Tab", "Toggle Sidebar"},
				{"Ctrl+c", "Force Quit"},
			},
		},
		{
			"Navigation",
			[][]string{
				{"↑/↓  j/k", "Navigate List"},
				{"Enter", "Select / Details"},
				{"/", "Filter Items"},
				{"Esc", "Go Back"},
			},
		},
		{
			"Actions",
			[][]string{
				{"r", "Refresh Data"},
				{"s", "Start Resource"},
				{"x", "Stop Resource"},
				{"h", "SSH Connect"},
				{"l", "Log Tailing"},
			},
		},
	}

	// Calculate column widths adaptively
	var columns []string
	for _, section := range sections {
		// Find max widths for this section
		maxKeyLen := len(section.title)
		maxDescLen := 0
		for _, item := range section.items {
			if len(item[0]) > maxKeyLen {
				maxKeyLen = len(item[0])
			}
			if len(item[1]) > maxDescLen {
				maxDescLen = len(item[1])
			}
		}

		// Add padding
		keyWidth := maxKeyLen + 3

		// Build column
		var col strings.Builder

		// Section header
		header := styles.TitleStyle.Copy().
			Foreground(styles.ColorHighlight).
			Bold(true).
			Underline(true).
			Render(section.title)
		col.WriteString(header + "\n\n")

		// Items
		for _, item := range section.items {
			key := lipgloss.NewStyle().
				Foreground(styles.ColorPrimary).
				Bold(true).
				Width(keyWidth).
				Render(item[0])

			desc := lipgloss.NewStyle().
				Foreground(styles.ColorText).
				Render(item[1])

			col.WriteString(key + desc + "\n")
		}

		columns = append(columns, col.String())
	}

	// Join columns with spacing
	content := lipgloss.JoinHorizontal(lipgloss.Top, columns[0], "   ", columns[1], "   ", columns[2])

	// Calculate dialog width based on content
	contentWidth := lipgloss.Width(content)
	dialogWidth := contentWidth + 8 // padding

	// Ensure it fits on screen
	if dialogWidth > width-4 {
		dialogWidth = width - 4
	}

	// Build final dialog
	title := styles.TitleStyle.Copy().
		Foreground(styles.ColorPrimary).
		Render("TGCP Help & Keybindings")

	// Get Banner
	banner := GetBanner()

	footer := styles.SubtleStyle.Render("Press ? or Esc to close")

	dialog := styles.BoxStyle.Copy().
		Width(dialogWidth).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorPrimary).
		Padding(1, 2).
		Render(lipgloss.JoinVertical(lipgloss.Center,
			banner,
			"\n",
			title,
			"",
			content,
			"",
			footer,
		))

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, dialog)
}
