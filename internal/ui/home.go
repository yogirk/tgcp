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
	screen := lipgloss.JoinVertical(lipgloss.Top, content, statusBar)

	// 3. Toast Overlay (if active)
	if m.Toast != nil && !m.Toast.IsExpired() {
		toastView := m.Toast.View()
		// Position toast at bottom-right, above status bar
		screen = lipgloss.JoinVertical(lipgloss.Top,
			content,
			lipgloss.PlaceHorizontal(m.Width, lipgloss.Right, toastView),
			statusBar,
		)
	}

	// 4. Loading Spinner (if active) - show inline at top of content
	if m.Spinner.IsActive() {
		spinnerView := m.Spinner.View()
		screen = lipgloss.JoinVertical(lipgloss.Top, spinnerView, content, statusBar)
	}

	// 5. Overlays (Command Palette)
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
	infoBox := styles.PrimaryBoxStyle.Copy().
		Render(userInfo)

	// Menu
	menu := m.HomeMenu.View()

	// Minimal navigation hint
	hints := styles.SubtleStyle.Render("↑/↓ navigate   Space expand/collapse   Enter select   ? help   : palette")

	// Version info
	versionText := styles.SubtleStyle.Render(m.Version.FormatVersion())

	// Update notification (if available)
	var updateNotice string
	if m.UpdateInfo != nil && m.UpdateInfo.Available {
		updateStyle := lipgloss.NewStyle().
			Foreground(styles.ColorSuccess).
			Bold(true)
		updateNotice = updateStyle.Render(
			fmt.Sprintf("Update available: %s -> %s", m.Version.FormatVersion(), "v"+m.UpdateInfo.LatestVersion),
		) + "\n" + styles.SubtleStyle.Render("Run: brew upgrade tgcp  or  visit github.com/yogirk/tgcp/releases")
	}

	// Layout: Center everything
	// We use lipgloss.Place to center vertically and horizontally

	// Combine components vertically
	bodyParts := []string{
		banner,
		"\n",
		infoBox,
		"\n",
		menu,
		"\n",
		hints,
		versionText,
	}

	// Add update notice if available
	if updateNotice != "" {
		bodyParts = append(bodyParts, "\n", updateNotice)
	}

	body := lipgloss.JoinVertical(lipgloss.Center, bodyParts...)

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
