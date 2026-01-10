package redis

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/yogirk/tgcp/internal/ui/components"
	"github.com/yogirk/tgcp/internal/styles"
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
		styles.BaseStyle.Foreground(statusColor).Render("🧠 "),
		styles.HeaderStyle.Render(fmt.Sprintf("Instance: %s", i.Name)),
	)

	details := fmt.Sprintf(
		"Display Name: %s\nLocation: %s\nTier: %s\nCapacity: %d GB\nVersion: %s\n\nHost: %s\nPort: %d\nNetwork: %s",
		i.DisplayName,
		i.Location,
		i.Tier,
		i.MemorySizeGb,
		i.RedisVersion,
		i.Host,
		i.Port,
		i.AuthorizedNetwork,
	)

	box := styles.BoxStyle.Copy().Width(80).Render(
		lipgloss.JoinVertical(lipgloss.Left, header, " ", details),
	)
	return box
}
