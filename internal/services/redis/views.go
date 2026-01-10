package redis

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/yogirk/tgcp/internal/styles"
	"github.com/yogirk/tgcp/internal/ui/components"
)

func (s *Service) View() string {
	if s.loading && len(s.instances) == 0 {
		return components.RenderSpinner("Loading Redis Instances...")
	}
	if s.err != nil {
		return components.RenderError(s.err, "Redis", "Instances")
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
			{Key: "Location", Value: i.Location},
			{Key: "Tier", Value: i.Tier},
			{Key: "Capacity", Value: fmt.Sprintf("%d GB", i.MemorySizeGb)},
			{Key: "Version", Value: i.RedisVersion},
			{Key: "Host", Value: i.Host},
			{Key: "Port", Value: fmt.Sprintf("%d", i.Port)},
			{Key: "Network", Value: i.AuthorizedNetwork},
		},
	})
	return lipgloss.JoinVertical(lipgloss.Left, breadcrumb, "", card)
}
