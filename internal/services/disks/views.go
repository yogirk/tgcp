package disks

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/yogirk/tgcp/internal/styles"
)

func (s *Service) renderDetailView() string {
	if s.selectedDisk == nil {
		return "No disk selected"
	}
	d := s.selectedDisk

	statusColor := styles.ColorSuccess
	if d.Status != "READY" {
		statusColor = styles.ColorWarning
	}

	// 1. Header with Status Bubble
	headerTitle := fmt.Sprintf("💾 Disk: %s (%s)", d.Name, d.Status)
	header := lipgloss.JoinHorizontal(lipgloss.Left,
		styles.BaseStyle.Foreground(statusColor).Render("● "),
		styles.HeaderStyle.Render(headerTitle),
	)

	// 2. Details Box
	detailsContent := fmt.Sprintf(
		"Zone: %s\nCreated: %s\nType: %s\nSize: %d GB\nSource Image: %s",
		d.Zone,
		"N/A", // Timestamp parsing left for polish
		d.ShortType(),
		d.SizeGb,
		d.SourceImage,
	)

	detailsBox := styles.BoxStyle.Copy().
		BorderForeground(styles.ColorSecondary).
		Width(80).
		Render(
			lipgloss.JoinVertical(lipgloss.Left,
				header,
				" ",
				detailsContent,
			),
		)

	// 3. Attachment Box
	var attachContent string
	if d.IsOrphan() {
		attachContent = styles.ErrorStyle.Render("🔴 ORPHAN: Not attached to any instance.")
	} else {
		var lines []string
		for _, u := range d.Users {
			parts := strings.Split(u, "/")
			instanceName := parts[len(parts)-1]
			lines = append(lines, fmt.Sprintf("• Instance: %s", instanceName))
		}
		attachContent = strings.Join(lines, "\n")
	}

	attachBox := styles.BoxStyle.Copy().
		BorderForeground(styles.ColorSubtext).
		Width(80).
		Render(
			lipgloss.JoinVertical(lipgloss.Left,
				styles.HeaderStyle.Render("🔗 Attachment"),
				attachContent,
			),
		)

	return lipgloss.JoinVertical(lipgloss.Left,
		detailsBox,
		"",
		attachBox,
	)
}
