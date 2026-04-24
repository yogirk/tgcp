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
	ColorWarning      = lipgloss.Color("214") // Orange — reserved for genuine warnings
	ColorError        = lipgloss.Color("196") // Red
	ColorInfo         = lipgloss.Color("45")  // Cyan — used for informational modes (filter, etc.)

	// Base Styles
	BaseStyle = lipgloss.NewStyle().
			Foreground(ColorTextPrimary)

	// ------------------------------------------------------------------
	// Box hierarchy
	// ------------------------------------------------------------------

	// PrimaryBoxStyle — main content cards (detail views, active panels).
	// Rounded border with accent color for visual prominence.
	PrimaryBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBrandAccent).
			Padding(1, 2)

	// SecondaryBoxStyle — supporting content (metadata, sections).
	// Normal border with subtle color, less prominent.
	SecondaryBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(ColorBorderSubtle).
				Padding(0, 1)

	// OverlayBoxStyle — modals, dialogs, dropdowns. Subtle rounded outline that
	// pairs well with action-specific border colors applied by the caller.
	OverlayBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorderSubtle).
			Padding(0, 1)

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

	// ------------------------------------------------------------------
	// Selection — two canonical states, used across sidebar, menu, palette
	// ------------------------------------------------------------------

	// SelectedActive — row is selected AND its list has focus.
	SelectedActive = lipgloss.NewStyle().
			Foreground(ColorBrandAccent).
			Bold(true).
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(ColorBrandAccent).
			Padding(0, 0, 0, 1)

	// SelectedBlur — row is selected but focus is elsewhere.
	SelectedBlur = lipgloss.NewStyle().
			Foreground(ColorBrandAccent).
			Padding(0, 0, 0, 2)

	UnselectedItemStyle = lipgloss.NewStyle().
				Foreground(ColorTextPrimary).
				Padding(0, 0, 0, 2)

	// SelectedItemStyle is retained as an alias for SelectedActive.
	SelectedItemStyle = SelectedActive

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
