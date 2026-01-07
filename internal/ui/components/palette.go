package components

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rk/tgcp/internal/core"
	"github.com/rk/tgcp/internal/styles"
)

type PaletteModel struct {
	TextInput textinput.Model
	Width     int
}

func NewPalette() PaletteModel {
	ti := textinput.New()
	ti.Placeholder = "Type a command..."
	ti.Prompt = "➜ "
	ti.CharLimit = 50
	ti.Width = 40
	ti.Focus() // Always focused when visible

	return PaletteModel{
		TextInput: ti,
	}
}

func (m PaletteModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m PaletteModel) Update(msg tea.Msg) (PaletteModel, tea.Cmd) {
	var cmd tea.Cmd
	m.TextInput, cmd = m.TextInput.Update(msg)
	return m, cmd
}

// Render renders the palette overlay using the provided navigation state
func (m PaletteModel) Render(nav core.NavigationModel, screenHeight int, banner string) string {
	// 1. Input Box
	// Using a "Search Bar" style (rounded, padded)
	inputBoxStyle := styles.FocusedBoxStyle.Copy().
		Width(60).
		Padding(1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorHighlight) // Pink/Focus color

	inputView := inputBoxStyle.Render(m.TextInput.View())

	// 2. Suggestions List
	var suggestionsView string
	if len(nav.Suggestions) > 0 {
		var lines []string
		for i, cmd := range nav.Suggestions {
			// Limit display to 8 items
			if i >= 8 {
				lines = append(lines, styles.SubtleStyle.Render(fmt.Sprintf("... and %d more", len(nav.Suggestions)-i)))
				break
			}

			// Render Item
			name := styles.LabelStyle.Render(cmd.Name)
			desc := styles.ValueStyle.Render(cmd.Description)

			// Use simple logic for layout
			// Name ........ Description
			// Or just "Name - Description"

			content := fmt.Sprintf("%-25s %s", name, desc)

			if i == nav.Selection {
				// Highlighted
				content = styles.SelectedItemStyle.Copy().
					Width(58). // Match box width approx (60 - padding)
					Render(content)
			} else {
				// Normal
				content = styles.UnselectedItemStyle.Copy().
					PaddingLeft(1).
					Render(content)
			}
			lines = append(lines, content)
		}
		suggestionsView = lipgloss.JoinVertical(lipgloss.Left, lines...)

		// Style the dropdown
		suggestionsView = styles.BoxStyle.Copy().
			Width(60).
			Border(lipgloss.RoundedBorder(), false, true, true, true). // No top border
			BorderForeground(styles.ColorSubtext).
			Render(suggestionsView)
	} else if m.TextInput.Value() != "" {
		// No matches
		suggestionsView = styles.BoxStyle.Copy().
			Width(60).
			Border(lipgloss.RoundedBorder(), false, true, true, true).
			BorderForeground(styles.ColorSubtext).
			Padding(0, 1).
			Render(styles.SubtleStyle.Render("No matching commands"))
	}

	// 3. Combine: Banner -> Buffer -> Input -> List
	// The banner is passed in.

	ui := lipgloss.JoinVertical(lipgloss.Center,
		banner,
		"\n", // Spacer between banner and search bar
		inputView,
		suggestionsView,
	)

	// 4. Center in Screen
	return lipgloss.Place(m.Width, screenHeight,
		lipgloss.Center, lipgloss.Center,
		ui,
	)
}
