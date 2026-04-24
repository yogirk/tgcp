package styles

import "github.com/charmbracelet/lipgloss"

// Spacing scale. All padding and margin values in the TUI should reference
// these constants so that rhythm stays consistent.
const (
	SpaceXS = 0
	SpaceS  = 1
	SpaceM  = 2
	SpaceL  = 4
)

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

	// Surface colors for status bar and headers
	ColorSurfaceBar    = lipgloss.Color("235") // Status bar background
	ColorSurfaceHeader = lipgloss.Color("237") // Header block background
	ColorTextOnBar     = lipgloss.Color("246") // Status bar foreground (≥4.5:1 on 235)

	// Accent palette — lifted from the Google-coloured banner so the rest of
	// the app can share the same visual register. Used to colour-code service
	// categories and light up category headers.
	ColorAccentRed    = lipgloss.Color("#DB4437")
	ColorAccentYellow = lipgloss.Color("#F4B400")
	ColorAccentGreen  = lipgloss.Color("#0F9D58")

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
			Padding(SpaceS, SpaceM)

	// SecondaryBoxStyle — supporting content (metadata, sections).
	// Normal border with subtle color, less prominent.
	SecondaryBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(ColorBorderSubtle).
				Padding(SpaceXS, SpaceS)

	// OverlayBoxStyle — modals, dialogs, dropdowns. Subtle rounded outline that
	// pairs well with action-specific border colors applied by the caller.
	OverlayBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorderSubtle).
			Padding(SpaceXS, SpaceS)

	// ------------------------------------------------------------------
	// Typography hierarchy — three tiers, one rule each
	// ------------------------------------------------------------------

	// HeaderStyle — page-level titles rendered as a solid bar.
	HeaderStyle = lipgloss.NewStyle().
			Foreground(ColorTextPrimary).
			Background(ColorSurfaceHeader).
			Bold(true).
			Padding(SpaceXS, SpaceS)

	// SectionStyle — card titles and in-box section headings.
	SectionStyle = lipgloss.NewStyle().
			Foreground(ColorBrandPrimary).
			Bold(true)

	// GroupStyle — muted uppercase dividers for list groups and categories.
	GroupStyle = lipgloss.NewStyle().
			Foreground(ColorTextMuted).
			Bold(true).
			PaddingLeft(SpaceS)

	// TitleStyle is retained as an alias for SectionStyle while callers migrate.
	TitleStyle = SectionStyle

	// ------------------------------------------------------------------
	// Sidebar
	// ------------------------------------------------------------------

	SidebarStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder(), false, true, false, false). // Right border only
			BorderForeground(ColorBorderSubtle).
			Padding(SpaceXS, SpaceS).
			Width(25)

	// ------------------------------------------------------------------
	// Selection — two canonical states, used across sidebar, menu, palette
	// ------------------------------------------------------------------

	// SelectedActive — row is selected AND its list has focus.
	SelectedActive = lipgloss.NewStyle().
			Foreground(ColorBrandAccent).
			Bold(true).
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(ColorBrandAccent).
			Padding(SpaceXS, SpaceXS, SpaceXS, SpaceS)

	// SelectedBlur — row is selected but focus is elsewhere.
	SelectedBlur = lipgloss.NewStyle().
			Foreground(ColorBrandAccent).
			Padding(SpaceXS, SpaceXS, SpaceXS, SpaceM)

	// UnselectedItemStyle — default row styling for list entries.
	UnselectedItemStyle = lipgloss.NewStyle().
				Foreground(ColorTextPrimary).
				Padding(SpaceXS, SpaceXS, SpaceXS, SpaceM)

	// SelectedItemStyle is retained as an alias for SelectedActive.
	SelectedItemStyle = SelectedActive

	// ------------------------------------------------------------------
	// Status bar
	// ------------------------------------------------------------------

	StatusBarStyle = lipgloss.NewStyle().
			Foreground(ColorTextOnBar).
			Background(ColorSurfaceBar).
			Padding(SpaceXS, SpaceS)

	// ------------------------------------------------------------------
	// Utility styles
	// ------------------------------------------------------------------

	LabelStyle = lipgloss.NewStyle().
			Foreground(ColorTextMuted).
			Bold(true).
			Width(10)

	ValueStyle = lipgloss.NewStyle().
			Foreground(ColorTextPrimary)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorError).
			Bold(true)

	SubtleStyle = lipgloss.NewStyle().
			Foreground(ColorTextMuted)

	SubtextStyle = SubtleStyle // Alias kept for existing callers

	SuccessStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess)

	WarningStyle = lipgloss.NewStyle().
			Foreground(ColorWarning)

	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorTextMuted).
			Italic(true)

	// Tabs
	ActiveTabStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder(), true, true, false, true).
			BorderForeground(ColorBrandAccent).
			Padding(SpaceXS, SpaceS).
			Bold(true).
			Foreground(ColorBrandAccent)

	InactiveTabStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder(), true, true, false, true).
				BorderForeground(ColorBorderSubtle).
				Padding(SpaceXS, SpaceS).
				Foreground(ColorTextMuted)
)
