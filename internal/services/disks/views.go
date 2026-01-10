package disks

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/yogirk/tgcp/internal/styles"
	"github.com/yogirk/tgcp/internal/ui/components"
)

func (s *Service) renderDetailView() string {
	if s.selectedDisk == nil {
		return "No disk selected"
	}
	d := s.selectedDisk

	breadcrumb := components.Breadcrumb(
		fmt.Sprintf("Project %s", s.projectID),
		s.Name(),
		"Disks",
		d.Name,
	)

	card := components.DetailCard(components.DetailCardOpts{
		Title: "Disk Details",
		Rows: []components.KeyValue{
			{Key: "Name", Value: d.Name},
			{Key: "Status", Value: components.RenderStatus(d.Status)},
			{Key: "Zone", Value: d.Zone},
			{Key: "Created", Value: "N/A"},
			{Key: "Type", Value: d.ShortType()},
			{Key: "Size", Value: fmt.Sprintf("%d GB", d.SizeGb)},
			{Key: "Source Image", Value: d.SourceImage},
		},
	})

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

	attachBox := components.DetailSection("Attachment", attachContent, styles.ColorBorderSubtle)

	return lipgloss.JoinVertical(lipgloss.Left,
		breadcrumb,
		"",
		card,
		"",
		attachBox,
	)
}
