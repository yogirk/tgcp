package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/yogirk/tgcp/internal/styles"
)

// View renders the application UI
func (m MainModel) View() string {
	// 1. Check for Help Overlay
	if m.ShowHelp {
		return HelpView(m.Width, m.Height)
	}

	// 2. Check for Start-up Error (Auth)
	if !m.AuthState.Authenticated {
		return renderAuthError(m)
	}

	var content string

	if m.ViewMode == ViewHome {
		content = renderLandingPage(m)
	} else {
		content = renderServiceLayout(m)
	}

	// Status Bar (Always Visible at bottom)
	statusBar := m.StatusBar.View()

	// Layout Content + Status Bar
	// But first, let's just make the root content
	screen := lipgloss.JoinVertical(lipgloss.Top, content, statusBar)

	// 3. Overlays (Command Palette)
	if m.Navigation.PaletteActive {
		// Overlay Palette on top of the entire screen
		// Note: Palette.Render uses lipgloss.Place to center itself in the given dimensions
		paletteView := m.Palette.Render(m.Navigation, m.Width, m.Height, GetBanner())

		// To truly "overlay" in TUI without clearing background is hard with just string concatenation.
		// However, lipgloss.Place will fill the screen with whitespace if we aren't careful.
		// A common trick is to just return only the palette view if we want it modal?
		// No, we want transparency or at least context.
		// But in simple TUI, rendering a "modal" often means rendering it *instead* of content, or
		// rendering it on top if the terminal supports cursor positioning hacks, but bubbletea 'View' returns a string.
		//
		// If we return just paletteView (which is centered), the rest is blank?
		// Palette.Render does `lipgloss.Place(..., ui)`. If whitespace is handled, it replaces screen.
		// Let's try just returning the palette view for focus, as it's a modal task.
		// It's cleaner than trying to merge strings.
		return paletteView
	}

	return screen
}

// renderLandingPage renders the central home screen
func renderLandingPage(m MainModel) string {

	// Banner
	banner := GetBanner()

	// User Info Box
	userInfo := fmt.Sprintf(
		"👤 User: %s    📁 Project: %s",
		m.AuthState.UserEmail,
		m.AuthState.ProjectID,
	)
	infoBox := styles.BoxStyle.Copy().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorSecondary).
		Padding(1).
		Render(userInfo)

	// Menu
	menu := m.HomeMenu.View()

	// Navigation Hints
	col1 := lipgloss.JoinVertical(lipgloss.Left,
		styles.SubtleStyle.Render("Global"),
		"q      Quit / Back",
		"?      Toggle Help",
		":      Cmd Palette",
		"/      Filter List",
	)

	col2 := lipgloss.JoinVertical(lipgloss.Left,
		styles.SubtleStyle.Render("Navigation"),
		"↑/↓    Move Cursor",
		"← / →  Focus Panes",
		"Enter  Select / Detail",
		"Tab    Toggle Sidebar",
		"Esc    Go Back",
	)

	col3 := lipgloss.JoinVertical(lipgloss.Left,
		styles.SubtleStyle.Render("Actions"),
		"r      Refresh Data",
		"s      Start Resource",
		"x      Stop Resource",
		"h      SSH Connect",
	)

	hints := styles.BoxStyle.Copy().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorSecondary).
		Padding(1, 2).
		Render(lipgloss.JoinHorizontal(lipgloss.Top,
			col1,
			"    ", // spacer
			col2,
			"    ", // spacer
			col3,
		))

	// Layout: Center everything
	// We use lipgloss.Place to center vertically and horizontally

	// Combine components vertically
	body := lipgloss.JoinVertical(
		lipgloss.Center,
		banner,
		"\n",
		infoBox,
		"\n",
		menu,
		"\n",
		hints,
	)

	return lipgloss.Place(
		m.Width, m.Height-1, // -1 for status bar space if needed
		lipgloss.Center, lipgloss.Center,
		body,
	)
}

// renderServiceLayout renders the sidebar + service content
func renderServiceLayout(m MainModel) string {
	// Left: Sidebar (if visible)
	leftPanel := m.Sidebar.View()

	// Right: Service View
	var rightPanel string
	if m.CurrentSvc != nil {
		rightPanel = m.CurrentSvc.View()
	} else {
		rightPanel = styles.BaseStyle.Render("Select a service from the sidebar.")
	}

	// Calculate Right Panel Width
	sidebarWidth := lipgloss.Width(leftPanel)
	rightPanelWidth := m.Width - sidebarWidth

	// Apply style to right panel to fill space
	rightPanelStyle := lipgloss.NewStyle().
		Width(rightPanelWidth).
		Height(m.Sidebar.Height).
		Padding(1)

	rightPanel = rightPanelStyle.Render(rightPanel)

	return lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)
}
