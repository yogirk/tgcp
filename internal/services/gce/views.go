package gce

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/rk/tgcp/internal/styles"
)

// renderDetailView renders the details of a single instance
func (s *Service) renderDetailView() string {
	if s.selectedInstance == nil {
		return "No instance selected"
	}
	i := s.selectedInstance

	doc := strings.Builder{}

	// Breadcrumb
	doc.WriteString(styles.SubtleStyle.Render(fmt.Sprintf("Google Compute Engine > Instances > %s", i.Name)))
	doc.WriteString("\n\n")

	// Instance Details Section
	detailsContent := fmt.Sprintf(`
Name:           %s
Status:         %s
Zone:           %s
Machine Type:   %s
Internal IP:    %s
External IP:    %s
`,
		i.Name,
		renderStatus(i.State),
		i.Zone,
		i.MachineType,
		i.InternalIP,
		i.ExternalIP,
	)

	detailsBox := styles.BoxStyle.Copy().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorSecondary).
		Padding(1).
		Width(80).
		Render(detailsContent)

	doc.WriteString(lipgloss.JoinVertical(lipgloss.Top,
		styles.LabelStyle.Render("╭─ Instance Details ─────────────────────────────────────────────╮"),
		detailsBox,
	))

	doc.WriteString("\n\n")

	// Actions Bar
	actions := "s Start | x Stop | h SSH | q Back"
	doc.WriteString(styles.HelpStyle.Render(actions))

	return doc.String()
}

func renderStatus(state InstanceState) string {
	str := string(state)
	switch state {
	case StateRunning:
		return styles.SuccessStyle.Render("🟢 " + str)
	case StateStopped, StateTerminated:
		return styles.WarningStyle.Render("🟡 " + str)
	default:
		return styles.ErrorStyle.Render("🔴 " + str)
	}
}

// renderConfirmation renders a confirmation dialog
func (s *Service) renderConfirmation() string {
	if s.selectedInstance == nil {
		return "Error: No instance selected"
	}

	actionTitle := "Confirm Action"
	actionText := ""

	if s.pendingAction == "start" {
		actionText = fmt.Sprintf("Are you sure you want to START instance %s?", styles.TitleStyle.Render(s.selectedInstance.Name))
	} else {
		actionText = fmt.Sprintf("Are you sure you want to STOP instance %s?", styles.TitleStyle.Render(s.selectedInstance.Name))
	}

	prompt := styles.HelpStyle.Render("y (Confirm) / n (Cancel)")

	content := lipgloss.JoinVertical(lipgloss.Center,
		styles.WarningStyle.Render(actionTitle),
		"\n",
		actionText,
		"\n",
		prompt,
	)

	dialog := styles.BoxStyle.Copy().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorWarning).
		Padding(1, 4).
		Render(content)

	// Center the dialog in the view
	// Since we don't have direct access to View's height/width here easily without passing it,
	// we will just return the dialog and let the parent layout handle it, or use Place if we tracked size.
	// But `Service` doesn't strictly track its own dimensions unless we stored them in WindowSizeMsg.
	// For now, let's just return the dialog, it will be top-left aligned in the content area which is fine.
	// Or we can simple padding to top/left to center it a bit.

	return lipgloss.Place(
		80, 20, // Approximate box
		lipgloss.Center, lipgloss.Center,
		dialog,
	)
}

// renderListView renders the main instance table
func (s *Service) renderListView() string {
	doc := strings.Builder{}

	// Filter Bar
	if s.filtering || s.filterInput.Value() != "" {
		doc.WriteString(s.filterInput.View())
		doc.WriteString("\n")
	}

	doc.WriteString(styles.BaseStyle.Render(s.table.View()))
	return doc.String()
}
