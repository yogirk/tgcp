package spanner

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/yogirk/tgcp/internal/ui/components"
)

func (s *Service) View() string {
	if s.loading && len(s.instances) == 0 {
		return components.RenderSpinner("Loading Spanner Instances...")
	}
	if s.err != nil {
		return components.RenderError(s.err, "Spanner", "Instances")
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

	breadcrumb := components.Breadcrumb(
		fmt.Sprintf("Project %s", s.projectID),
		s.Name(),
		"Instances",
		i.Name,
	)

	capacity := fmt.Sprintf("%d Nodes", i.NodeCount)
	if i.NodeCount == 0 {
		capacity = fmt.Sprintf("%d Processing Units", i.ProcessingUnits)
	}

	card := components.DetailCard(components.DetailCardOpts{
		Title: "Instance Details",
		Rows: []components.KeyValue{
			{Key: "Name", Value: i.Name},
			{Key: "Status", Value: components.RenderStatus(i.State)},
			{Key: "Display Name", Value: i.DisplayName},
			{Key: "Configuration", Value: i.Config},
			{Key: "Capacity", Value: capacity},
		},
	})
	return lipgloss.JoinVertical(lipgloss.Left, breadcrumb, "", card)
}
