package gce

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/yogirk/tgcp/internal/styles"
	"github.com/yogirk/tgcp/internal/ui/components"
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
	// Calculate Total Disk Size
	var totalDisk int64
	for _, d := range i.Disks {
		totalDisk += d.SizeGB
	}

	// Calculate Age
	// Simple duration format: Xd Yh
	age := time.Since(i.CreationTime)
	days := int(age.Hours() / 24)
	ageStr := fmt.Sprintf("%d days ago", days)
	if days == 0 {
		hours := int(age.Hours())
		ageStr = fmt.Sprintf("%d hours ago", hours)
	}

	detailsContent := fmt.Sprintf(`
Name:           %s
Status:         %s
Zone:           %s
Machine Type:   %s
OS Image:       %s
Disk Size:      %d GB
Created:        %s
Estimated Cost: %s
Internal IP:    %s
External IP:    %s
`,
		i.Name,
		renderStatus(i.State),
		i.Zone,
		i.MachineType,
		i.OSImage,
		totalDisk,
		ageStr,
		EstimateCost(i.MachineType, i.Zone, i.Disks),
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
		return styles.SuccessStyle.Render("● " + str)
	case StateStopped, StateTerminated:
		label := str
		if state == StateTerminated {
			label = "STOP"
		}
		return styles.ErrorStyle.Render("● " + label)
	case StateProvisioning, StateStaging, StateStopping, StateSuspending, StateRepairing:
		return styles.WarningStyle.Render("● " + str)
	default:
		return styles.SubtextStyle.Render("● " + str)
	}
}

// renderConfirmation renders a confirmation dialog
func (s *Service) renderConfirmation() string {
	if s.selectedInstance == nil {
		return "Error: No instance selected"
	}

	return components.RenderConfirmation(s.pendingAction, s.selectedInstance.Name, "instance")
}

// renderListView renders the main instance table
func (s *Service) renderListView() string {
	doc := strings.Builder{}

	// Filter Bar
	doc.WriteString(s.filter.View())
	doc.WriteString("\n")

	doc.WriteString(styles.BaseStyle.Render(s.table.View()))
	return doc.String()
}
