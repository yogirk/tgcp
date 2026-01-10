package bigtable

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/yogirk/tgcp/internal/ui/components"
	"github.com/yogirk/tgcp/internal/styles"
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
	if s.filter.IsActive() || s.filter.Value() != "" {
		content.WriteString(s.filter.View())
		content.WriteString("\n")
	}
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

	header := lipgloss.JoinHorizontal(lipgloss.Left,
		styles.BaseStyle.Foreground(statusColor).Render("🥞 "),
		styles.HeaderStyle.Render(fmt.Sprintf("Instance: %s", i.Name)),
	)

	details := fmt.Sprintf(
		"Display Name: %s\nType: %s\nState: %s\nProject: %s",
		i.DisplayName,
		i.Type,
		i.State,
		i.ProjectID,
	)

	box := styles.BoxStyle.Copy().Width(80).Render(
		lipgloss.JoinVertical(lipgloss.Left, header, " ", details),
	)

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

	clusterBox := styles.BoxStyle.Copy().
		BorderForeground(styles.ColorSubtext).
		Width(80).
		Render(
			lipgloss.JoinVertical(lipgloss.Left,
				styles.HeaderStyle.Render("Clusters"),
				clusterContent,
			),
		)

	return lipgloss.JoinVertical(lipgloss.Left,
		box,
		"",
		clusterBox,
	)
}
