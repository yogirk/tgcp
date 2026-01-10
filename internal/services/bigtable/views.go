package bigtable

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/yogirk/tgcp/internal/styles"
	"github.com/yogirk/tgcp/internal/ui/components"
)

func (s *Service) View() string {
	if s.loading && len(s.instances) == 0 {
		return components.RenderSpinner("Loading Bigtable Instances...")
	}
	if s.err != nil {
		return components.RenderError(s.err, "Bigtable", "Instances")
	}

	if s.viewState == ViewDetail {
		return s.renderDetailView()
	}

	// Filter Bar
	var content strings.Builder
	content.WriteString(components.Breadcrumb(
		fmt.Sprintf("Project %s", s.projectID),
		s.Name(),
		"Instances",
	))
	content.WriteString("\n")
	content.WriteString(s.filter.View())
	content.WriteString("\n")
	content.WriteString(s.table.View())
	return content.String()
}

func (s *Service) renderDetailView() string {
	i := s.selectedInstance
	if i == nil {
		return ""
	}

	statusColor := styles.ColorSuccess
	if i.State != "READY" {
		statusColor = styles.ColorWarning
	}

	breadcrumb := components.Breadcrumb(
		fmt.Sprintf("Project %s", s.projectID),
		s.Name(),
		"Instances",
		i.Name,
	)

	card := components.DetailCard(components.DetailCardOpts{
		Title: "Instance Details",
		Rows: []components.KeyValue{
			{Key: "Name", Value: i.Name},
			{Key: "Status", Value: styles.BaseStyle.Foreground(statusColor).Render(i.State)},
			{Key: "Display Name", Value: i.DisplayName},
			{Key: "Type", Value: i.Type},
			{Key: "Project", Value: i.ProjectID},
		},
	})

	// Clusters
	clusterContent := components.RenderSpinner("Loading clusters...")
	if s.clusters != nil {
		if len(s.clusters) == 0 {
			clusterContent = "No clusters found."
		} else {
			var lines []string
			for _, c := range s.clusters {
				line := fmt.Sprintf(
					"• %s (%s): %d Nodes, %s [%s]",
					c.Name,
					c.Zone,
					c.ServeNodes,
					c.StorageType,
					c.State,
				)
				lines = append(lines, line)
			}
			clusterContent = strings.Join(lines, "\n")
		}
	}

	clusterBox := components.DetailSection("Clusters", clusterContent, styles.ColorBorderSubtle)

	return lipgloss.JoinVertical(lipgloss.Left,
		breadcrumb,
		"",
		card,
		"",
		clusterBox,
	)
}
