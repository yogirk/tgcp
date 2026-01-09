package spanner

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/rk/tgcp/internal/styles"
)

func (s *Service) View() string {
	if s.loading && len(s.instances) == 0 {
		return "Loading Spanner Instances..."
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
	i := s.selectedInstance
	if i == nil {
		return ""
	}

	statusColor := styles.ColorSuccess
	if i.State != "READY" {
		statusColor = styles.ColorWarning
	}

	header := lipgloss.JoinHorizontal(lipgloss.Left,
		styles.BaseStyle.Foreground(statusColor).Render("🌍 "),
		styles.HeaderStyle.Render(fmt.Sprintf("Instance: %s", i.Name)),
	)

	capacity := fmt.Sprintf("%d Nodes", i.NodeCount)
	if i.NodeCount == 0 {
		capacity = fmt.Sprintf("%d Processing Units", i.ProcessingUnits)
	}

	details := fmt.Sprintf(
		"Display Name: %s\nConfiguration: %s\nCapacity: %s\nState: %s",
		i.DisplayName,
		i.Config,
		capacity,
		i.State,
	)

	box := styles.BoxStyle.Copy().Width(80).Render(
		lipgloss.JoinVertical(lipgloss.Left, header, " ", details),
	)
	return box
}
