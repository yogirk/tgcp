package cloudsql

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/yogirk/tgcp/internal/ui/components"
	"github.com/yogirk/tgcp/internal/styles"
)

func (s *Service) renderDetailView() string {
	if s.selectedInstance == nil {
		return "No instance selected"
	}
	i := s.selectedInstance

	doc := strings.Builder{}

	// Breadcrumb
	doc.WriteString(styles.SubtleStyle.Render(fmt.Sprintf("Cloud SQL > Instances > %s", i.Name)))
	doc.WriteString("\n\n")

	// Details
	content := fmt.Sprintf(`
Name:              %s
State:             %s
Database Version:  %s
Region:            %s
Tier:              %s
Storage (GB):      %d
Auto Backup:       %v
Activation Policy: %s
Primary IP:        %s
Connection Name:   %s
`,
		i.Name,
		renderState(i.State),
		i.DatabaseVersion,
		i.Region,
		i.Tier,
		i.StorageGB,
		i.AutoBackup,
		i.Activation,
		i.PrimaryIP,
		i.ConnectionName,
	)

	box := styles.BoxStyle.Copy().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorSecondary).
		Padding(1).
		Width(80).
		Render(content)

	doc.WriteString(lipgloss.JoinVertical(lipgloss.Top,
		styles.LabelStyle.Render("╭─ Instance Details ─────────────────────────────────────────────╮"),
		box,
	))

	doc.WriteString("\n\n")
	doc.WriteString(styles.HelpStyle.Render("s Start | x Stop | q Back"))

	return doc.String()
}

func (s *Service) renderConfirmation() string {
	if s.selectedInstance == nil {
		return "Error: No instance selected"
	}

	return components.RenderConfirmation(s.pendingAction, s.selectedInstance.Name, "instance")
}

func renderState(state InstanceState) string {
	str := string(state)
	switch state {
	case StateRunnable:
		return styles.SuccessStyle.Render("● " + str)
	case StateSuspended, StatePending:
		return styles.WarningStyle.Render("● " + str)
	case StateFailed:
		return styles.ErrorStyle.Render("● " + str)
	default:
		return str
	}
}
