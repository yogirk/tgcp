package dataproc

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/yogirk/tgcp/internal/styles"
	"github.com/yogirk/tgcp/internal/ui/components"
)

func (s *Service) View() string {
	if s.loading && len(s.clusters) == 0 {
		return components.RenderSpinner("Loading Dataproc Clusters (us-central1)...")
	}
	if s.err != nil {
		return components.RenderError(s.err, "Dataproc", "Clusters")
	}

	if s.viewState == ViewDetail {
		return s.renderDetailView()
	}

	// Filter Bar
	var content strings.Builder
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

	statusColor := styles.ColorSuccess
	if c.Status != "RUNNING" {
		statusColor = styles.ColorWarning
	}

	header := lipgloss.JoinHorizontal(lipgloss.Left,
		styles.BaseStyle.Foreground(statusColor).Render("🐘 "),
		styles.HeaderStyle.Render(fmt.Sprintf("Cluster: %s", c.Name)),
	)

	workers := fmt.Sprintf("%d x %s", c.WorkerCount, c.WorkerMachine)

	details := fmt.Sprintf(
		"Status: %s\nDefault Region: %s\nZone: %s\n\nMaster: %s\nWorkers: %s",
		c.Status,
		DefaultRegion,
		c.Zone,
		c.MasterMachine,
		workers,
	)

	box := styles.BoxStyle.Copy().Width(80).Render(
		lipgloss.JoinVertical(lipgloss.Left, header, " ", details),
	)
	return box
}
