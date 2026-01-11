package dataproc

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/yogirk/tgcp/internal/ui/components"
)

func (s *Service) View() string {
	if s.err != nil {
		return components.RenderError(s.err, "Dataproc", "Clusters")
	}

	// Show spinner while loading
	if s.spinner.IsActive() {
		return s.spinner.View()
	}

	if s.viewState == ViewDetail {
		return s.renderDetailView()
	}

	// Filter Bar
	var content strings.Builder
	content.WriteString(components.Breadcrumb(
		fmt.Sprintf("Project %s", s.projectID),
		s.Name(),
		"Clusters",
	))
	content.WriteString("\n")
	content.WriteString(s.filter.View())
	content.WriteString("\n")
	content.WriteString(s.table.View())
	return content.String()
}

func (s *Service) renderDetailView() string {
	c := s.selectedCluster
	if c == nil {
		return ""
	}

	breadcrumb := components.Breadcrumb(
		fmt.Sprintf("Project %s", s.projectID),
		s.Name(),
		"Clusters",
		c.Name,
	)

	workers := fmt.Sprintf("%d x %s", c.WorkerCount, c.WorkerMachine)

	card := components.DetailCard(components.DetailCardOpts{
		Title: "Cluster Details",
		Rows: []components.KeyValue{
			{Key: "Name", Value: c.Name},
			{Key: "Status", Value: components.RenderStatus(c.Status)},
			{Key: "Region", Value: DefaultRegion},
			{Key: "Zone", Value: c.Zone},
			{Key: "Master", Value: c.MasterMachine},
			{Key: "Workers", Value: workers},
		},
	})
	return lipgloss.JoinVertical(lipgloss.Left, breadcrumb, "", card)
}
