package dataflow

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/rk/tgcp/internal/styles"
)

func (s *Service) View() string {
	if s.loading && len(s.jobs) == 0 {
		return "Loading Dataflow Jobs..."
	}
	if s.err != nil {
		return fmt.Sprintf("Error: %v", s.err)
	}

	if s.viewState == ViewDetail {
		return s.renderDetailView()
	}

	return s.table.View()
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

	header := lipgloss.JoinHorizontal(lipgloss.Left,
		styles.BaseStyle.Foreground(statusColor).Render("🌊 "),
		styles.HeaderStyle.Render(fmt.Sprintf("Job: %s", j.Name)),
	)

	details := fmt.Sprintf(
		"ID: %s\nType: %s\nState: %s\nLocation: %s\nCreated: %s",
		j.ID,
		strings.Replace(j.Type, "JOB_TYPE_", "", 1),
		cleanState,
		j.Location,
		j.CreateTime, // ISO string
	)

	box := styles.BoxStyle.Copy().Width(80).Render(
		lipgloss.JoinVertical(lipgloss.Left, header, " ", details),
	)
	return box
}
