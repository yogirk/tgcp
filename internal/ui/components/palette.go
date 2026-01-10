package components

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yogirk/tgcp/internal/core"
	"github.com/yogirk/tgcp/internal/styles"
)

type PaletteModel struct {
	TextInput textinput.Model
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
func (m PaletteModel) Render(nav core.NavigationModel, screenWidth, screenHeight int, banner string) string {
	if screenWidth <= 0 {
		screenWidth = 80
	}
	if screenHeight <= 0 {
		screenHeight = 24
	}

	boxWidth := screenWidth - 6
	if boxWidth > 72 {
		boxWidth = 72
	}
	if boxWidth < 40 {
		boxWidth = 40
	}

	// 1. Input Box
	// Using a "Search Bar" style (rounded, padded)
	inputBoxStyle := styles.FocusedBoxStyle.Copy().
		Width(boxWidth).
		Padding(1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorBrandAccent) // Pink/Focus color

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
					Width(boxWidth - 2). // Match box width approx (padding)
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
			Width(boxWidth).
			Border(lipgloss.RoundedBorder(), false, true, true, true). // No top border
			BorderForeground(styles.ColorTextMuted).
			Render(suggestionsView)
	} else if m.TextInput.Value() != "" {
		// No matches
		suggestionsView = styles.BoxStyle.Copy().
			Width(boxWidth).
			Border(lipgloss.RoundedBorder(), false, true, true, true).
			BorderForeground(styles.ColorTextMuted).
			Padding(0, 1).
			Render(styles.SubtleStyle.Render("No matching commands"))
	}

	helpHint := styles.SubtleStyle.Render("Esc:Cancel  Enter:Run  ↑/↓:Select")

	// 3. Combine: Banner -> Buffer -> Input -> List
	// The banner is passed in.

	ui := lipgloss.JoinVertical(lipgloss.Center,
		banner,
		"\n", // Spacer between banner and search bar
		inputView,
		suggestionsView,
		"\n",
		helpHint,
	)

	// 4. Center in Screen without backdrop to avoid ghosting/shadows
	return lipgloss.Place(screenWidth, screenHeight,
		lipgloss.Center, lipgloss.Center,
		ui,
	)
}
