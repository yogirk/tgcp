package styles

import "github.com/charmbracelet/lipgloss"

var (
	// Semantic colors
	ColorBrandPrimary = lipgloss.Color("39")  // GCP Blue (brighter)
	ColorBrandAccent  = lipgloss.Color("75")  // Light Blue (for highlights)
	ColorTextPrimary  = lipgloss.Color("252") // Near white
	ColorTextMuted    = lipgloss.Color("243") // Muted grey (more contrast from primary)
	ColorBorderSubtle = lipgloss.Color("240") // Subtle border (more visible)
	ColorSuccess      = lipgloss.Color("42")  // Green
	ColorWarning      = lipgloss.Color("214") // Orange
	ColorError        = lipgloss.Color("196") // Red
	ColorInfo         = lipgloss.Color("45")  // Cyan

	// Base Styles
	BaseStyle = lipgloss.NewStyle().
			Foreground(ColorTextPrimary)

	// Component Styles - Box Hierarchy
	// BoxStyle is the base (kept for backwards compatibility)
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorderSubtle).
			Padding(0, 1)

	// PrimaryBoxStyle - for main content cards (detail views, active panels)
	// Rounded border with accent color for visual prominence
	PrimaryBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBrandAccent).
			Padding(1, 2)

	// SecondaryBoxStyle - for supporting content (metadata, sections)
	// Normal border with subtle color, less prominent
	SecondaryBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(ColorBorderSubtle).
			Padding(0, 1)

	FocusedBoxStyle = BoxStyle.Copy().
			BorderForeground(ColorBrandPrimary)

	HeaderStyle = lipgloss.NewStyle().
			Foreground(ColorTextPrimary).
			Background(lipgloss.Color("237")).
			Bold(true).
			Padding(0, 1)

	TitleStyle = lipgloss.NewStyle().
			Foreground(ColorBrandPrimary).
			Bold(true)

	// Sidebar Styles
	SidebarStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder(), false, true, false, false). // Right border only
			BorderForeground(ColorBorderSubtle).
			Padding(0, 1).
			Width(25) // Fixed width for sidebar

	SelectedItemStyle = lipgloss.NewStyle().
				Foreground(ColorBrandAccent).
				Bold(true).
				Border(lipgloss.NormalBorder(), false, false, false, true). // Left border
				BorderForeground(ColorBrandAccent).
				Padding(0, 0, 0, 1)

	UnselectedItemStyle = lipgloss.NewStyle().
				Foreground(ColorTextPrimary).
				Padding(0, 0, 0, 2)

	// Status Bar Styles
	StatusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Background(lipgloss.Color("235")).
			Padding(0, 1)

	// Generic Styles
	LabelStyle = lipgloss.NewStyle().
			Foreground(ColorTextMuted).
			Bold(true).
			Width(10)

	ValueStyle = lipgloss.NewStyle().
			Foreground(ColorTextPrimary)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorError).
			Bold(true)

	// New Styles
	SubtleStyle = lipgloss.NewStyle().
			Foreground(ColorTextMuted)

	SubtextStyle = SubtleStyle // Alias for SubtextStyle used in views

	SuccessStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess)

	WarningStyle = lipgloss.NewStyle().
			Foreground(ColorWarning)

	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorTextMuted).
			Italic(true)

	// Tab Styles
	ActiveTabStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder(), true, true, false, true).
			BorderForeground(ColorBrandAccent).
			Padding(0, 1).
			Bold(true).
			Foreground(ColorBrandAccent)

	InactiveTabStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder(), true, true, false, true).
				BorderForeground(ColorBorderSubtle).
				Padding(0, 1).
				Foreground(ColorTextMuted)
)
