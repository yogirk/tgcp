package dataflow

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/yogirk/tgcp/internal/ui/components"
)

func (s *Service) View() string {
	if s.err != nil {
		return components.RenderError(s.err, "Dataflow", "Jobs")
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
		"Jobs",
	))
	content.WriteString("\n")
	content.WriteString(s.filter.View())
	content.WriteString("\n")
	content.WriteString(s.table.View())
	return content.String()
}

func (s *Service) renderDetailView() string {
	j := s.selectedJob
	if j == nil {
		return ""
	}

	breadcrumb := components.Breadcrumb(
		fmt.Sprintf("Project %s", s.projectID),
		s.Name(),
		"Jobs",
		j.Name,
	)

	card := components.DetailCard(components.DetailCardOpts{
		Title: "Job Details",
		Rows: []components.KeyValue{
			{Key: "Name", Value: j.Name},
			{Key: "ID", Value: j.ID},
			{Key: "Type", Value: strings.Replace(j.Type, "JOB_TYPE_", "", 1)},
			{Key: "State", Value: components.RenderStatus(j.State)},
			{Key: "Location", Value: j.Location},
			{Key: "Created", Value: j.CreateTime},
		},
	})
	return lipgloss.JoinVertical(lipgloss.Left, breadcrumb, "", card)
}
