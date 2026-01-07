package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/rk/tgcp/internal/styles"
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

	// Overlays (Command Palette)
	if m.Navigation.PaletteActive {
		// Render palette at bottom
		palette := styles.BoxStyle.Render("Command Palette: " + m.StatusBar.Message)
		content = lipgloss.JoinVertical(lipgloss.Top, content, palette)
	} else {
		// Render Status Bar
		content = lipgloss.JoinVertical(lipgloss.Top, content, m.StatusBar.View())
	}

	return content
}

// renderLandingPage renders the central home screen
func renderLandingPage(m MainModel) string {

	// Banner
	banner := GetBanner()

	// User Info Box
	userInfo := fmt.Sprintf(
		"👤 User: %s\n📁 Project: %s",
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
	hints := styles.HeaderStyle.Render("⌨️  Navigation: Enter (Select) | q (Quit)")

	// Layout: Center everything
	// We use lipgloss.Place to center vertically and horizontally

	// Combine components vertically
	body := lipgloss.JoinVertical(
		lipgloss.Center,
		banner,
		"\n",
		infoBox,
		"\n\n",
		menu,
		"\n\n",
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
