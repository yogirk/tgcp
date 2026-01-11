package components

import (
	"fmt"
	"strings"

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
	ti.Placeholder = "Search services, actions..."
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

// highlightMatches renders text with matched characters highlighted
// matchedIndexes are positions in the full "Name Description" string
// We only highlight the name portion for cleaner display
func highlightMatches(name, description string, matchedIndexes []int) (string, string) {
	// Build a set of matched indexes for O(1) lookup
	matchSet := make(map[int]bool)
	for _, idx := range matchedIndexes {
		matchSet[idx] = true
	}

	// Style for highlighted (matched) characters
	highlightStyle := lipgloss.NewStyle().
		Foreground(styles.ColorBrandAccent).
		Bold(true)

	// Style for normal characters in name
	nameStyle := lipgloss.NewStyle().
		Foreground(styles.ColorTextPrimary).
		Bold(true)

	// Build highlighted name
	var nameBuilder strings.Builder
	for i, r := range name {
		char := string(r)
		if matchSet[i] {
			nameBuilder.WriteString(highlightStyle.Render(char))
		} else {
			nameBuilder.WriteString(nameStyle.Render(char))
		}
	}

	// Description uses muted style, highlight matches there too
	descStyle := styles.SubtleStyle
	var descBuilder strings.Builder
	nameLen := len(name) + 1 // +1 for the space between name and description

	for i, r := range description {
		char := string(r)
		if matchSet[nameLen+i] {
			descBuilder.WriteString(highlightStyle.Render(char))
		} else {
			descBuilder.WriteString(descStyle.Render(char))
		}
	}

	return nameBuilder.String(), descBuilder.String()
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
	// Determine if we have a dropdown (suggestions or "no matches")
	hasDropdown := len(nav.Suggestions) > 0 || m.TextInput.Value() != ""

	// Build input box style - seamless with dropdown when present
	var inputBoxStyle lipgloss.Style
	if hasDropdown {
		// No bottom border - connects seamlessly with dropdown
		inputBoxStyle = lipgloss.NewStyle().
			Width(boxWidth).
			Padding(1).
			Border(lipgloss.RoundedBorder(), true, true, false, true). // No bottom
			BorderForeground(styles.ColorBrandAccent)
	} else {
		// Full border when no dropdown
		inputBoxStyle = lipgloss.NewStyle().
			Width(boxWidth).
			Padding(1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(styles.ColorBrandAccent)
	}

	inputView := inputBoxStyle.Render(m.TextInput.View())

	// 2. Suggestions List
	var suggestionsView string
	if len(nav.Suggestions) > 0 {
		var lines []string
		for i, match := range nav.Suggestions {
			// Limit display to 8 items
			if i >= 8 {
				lines = append(lines, styles.SubtleStyle.Render(fmt.Sprintf("... and %d more", len(nav.Suggestions)-i)))
				break
			}

			// Render Item with highlighted matches
			name, desc := highlightMatches(match.Name, match.Description, match.MatchedIndexes)

			// Layout: Name (padded) Description
			// Use lipgloss width for proper padding with ANSI codes
			nameWidth := 25
			nameRendered := lipgloss.NewStyle().Width(nameWidth).Render(name)
			content := nameRendered + " " + desc

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

		// Style the dropdown - no top border, same accent color as input
		suggestionsView = styles.BoxStyle.Copy().
			Width(boxWidth).
			Border(lipgloss.RoundedBorder(), false, true, true, true). // No top border
			BorderForeground(styles.ColorBrandAccent).                 // Match input border color
			Render(suggestionsView)
	} else if m.TextInput.Value() != "" {
		// No matches - still connected to input
		suggestionsView = styles.BoxStyle.Copy().
			Width(boxWidth).
			Border(lipgloss.RoundedBorder(), false, true, true, true).
			BorderForeground(styles.ColorBrandAccent). // Match input border color
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
