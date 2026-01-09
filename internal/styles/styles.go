package styles

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	ColorPrimary   = lipgloss.Color("205") // Pink
	ColorSecondary = lipgloss.Color("62")  // Purple
	ColorText      = lipgloss.Color("252") // White-ish
	ColorSubtext   = lipgloss.Color("240") // Grey
	ColorSuccess   = lipgloss.Color("42")  // Green
	ColorWarning   = lipgloss.Color("214") // Orange
	ColorError     = lipgloss.Color("196") // Red
	ColorHighlight = lipgloss.Color("212") // Light Pink
	ColorAccent    = lipgloss.Color("39")  // Cyan/Blue for GKE headers

	// Base Styles
	BaseStyle = lipgloss.NewStyle().
			Foreground(ColorText)

	// Component Styles
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorSecondary).
			Padding(0, 1)

	FocusedBoxStyle = BoxStyle.Copy().
			BorderForeground(ColorPrimary)

	HeaderStyle = lipgloss.NewStyle().
			Foreground(ColorSubtext).
			Padding(0, 1)

	TitleStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)

	// Sidebar Styles
	SidebarStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder(), false, true, false, false). // Right border only
			BorderForeground(ColorSubtext).
			Padding(0, 1).
			Width(25) // Fixed width for sidebar

	SelectedItemStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary).
				Bold(true).
				Border(lipgloss.NormalBorder(), false, false, false, true). // Left border
				BorderForeground(ColorPrimary).
				Padding(0, 0, 0, 1)

	UnselectedItemStyle = lipgloss.NewStyle().
				Foreground(ColorText).
				Padding(0, 0, 0, 2)

	// Status Bar Styles
	StatusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Background(lipgloss.Color("235")).
			Padding(0, 1)

	// Generic Styles
	LabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("99")).
			Bold(true).
			Width(10)

	ValueStyle = lipgloss.NewStyle().
			Foreground(ColorText)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorError).
			Bold(true)

	// New Styles
	SubtleStyle = lipgloss.NewStyle().
			Foreground(ColorSubtext)

	SubtextStyle = SubtleStyle // Alias for SubtextStyle used in views

	SuccessStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess)

	WarningStyle = lipgloss.NewStyle().
			Foreground(ColorWarning)

	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorSubtext).
			Italic(true)

	// Tab Styles
	ActiveTabStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder(), true, true, false, true).
			BorderForeground(ColorPrimary).
			Padding(0, 1).
			Bold(true).
			Foreground(ColorPrimary)

	InactiveTabStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder(), true, true, false, true).
				BorderForeground(ColorSubtext).
				Padding(0, 1).
				Foreground(ColorSubtext)
)
