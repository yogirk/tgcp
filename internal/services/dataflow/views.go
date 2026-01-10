package dataflow

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/yogirk/tgcp/internal/styles"
	"github.com/yogirk/tgcp/internal/ui/components"
)

func (s *Service) View() string {
	if s.loading && len(s.jobs) == 0 {
		return components.RenderSpinner("Loading Dataflow Jobs...")
	}
	if s.err != nil {
		return components.RenderError(s.err, "Dataflow", "Jobs")
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

	cleanState := strings.Replace(j.State, "JOB_STATE_", "", 1)
	statusColor := styles.ColorSuccess
	if cleanState != "RUNNING" && cleanState != "DONE" {
		statusColor = styles.ColorWarning
	}
	if cleanState == "FAILED" || cleanState == "CANCELLED" {
		statusColor = styles.ColorError
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
			{Key: "State", Value: styles.BaseStyle.Foreground(statusColor).Render(cleanState)},
			{Key: "Location", Value: j.Location},
			{Key: "Created", Value: j.CreateTime},
		},
	})
	return lipgloss.JoinVertical(lipgloss.Left, breadcrumb, "", card)
}
