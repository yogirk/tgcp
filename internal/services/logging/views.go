package logging

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/yogirk/tgcp/internal/styles"
	"github.com/yogirk/tgcp/internal/ui/components"
)

// View renders the service UI
func (s *Service) View() string {
	if s.err != nil {
		return components.RenderError(s.err, s.Name(), "Logs")
	}

	// Show animated spinner while loading
	if s.spinner.IsActive() {
		return s.spinner.View()
	}

	if s.viewingDetail {
		return s.renderDetailView()
	}

	// Default: List View
	return s.renderListView()
}

// renderListView renders the compact log table
func (s *Service) renderListView() string {
	// Breadcrumb
	var crumbItems []string
	if s.heading != "" {
		crumbItems = []string{s.Name(), s.heading}
	} else {
		crumbItems = []string{s.Name(), "All Logs"}
	}

	breadcrumb := components.Breadcrumb(crumbItems...)

	// Page indicator
	pageInfo := ""
	if len(s.tokenStack) > 0 || s.nextPageToken != "" {
		page := len(s.tokenStack) + 1
		pageInfo = fmt.Sprintf("  Page %d", page)
		if s.nextPageToken != "" {
			pageInfo += "+"
		}
	}

	// Count indicator
	countInfo := fmt.Sprintf("(%d entries%s)", len(s.entries), pageInfo)
	countStyle := lipgloss.NewStyle().Foreground(styles.ColorTextMuted)

	header := lipgloss.JoinHorizontal(lipgloss.Left,
		breadcrumb,
		"  ",
		countStyle.Render(countInfo),
	)

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		s.table.View(),
	)
}

// renderDetailView shows full log entry details
func (s *Service) renderDetailView() string {
	if s.selectedEntry == nil {
		return "No entry selected"
	}

	e := s.selectedEntry

	breadcrumb := components.Breadcrumb(s.Name(), "Entry Detail")

	// Format the full payload with word wrap
	wrapWidth := s.width - 10
	if wrapWidth < 40 {
		wrapWidth = 80
	}

	// Metadata section
	metaRows := []components.KeyValue{
		{Key: "Timestamp", Value: e.Timestamp.Local().Format("2006-01-02 15:04:05.000")},
		{Key: "Severity", Value: e.Severity},
		{Key: "Resource", Value: fmt.Sprintf("%s / %s", e.ResourceType, e.ResourceName)},
		{Key: "Location", Value: e.Location},
		{Key: "Log Name", Value: e.LogName},
		{Key: "Insert ID", Value: e.InsertID},
	}

	metaCard := components.DetailCard(components.DetailCardOpts{
		Title: "Log Entry Metadata",
		Rows:  metaRows,
	})

	// Payload section with full content
	payloadStyle := lipgloss.NewStyle().
		Foreground(styles.ColorTextPrimary).
		Width(wrapWidth)

	payloadBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorBorderSubtle).
		Padding(1).
		Width(wrapWidth + 4)

	payloadContent := payloadStyle.Render(e.Payload)
	if e.FullPayload != "" && e.FullPayload != e.Payload {
		payloadContent = payloadStyle.Render(e.FullPayload)
	}

	payloadSection := payloadBox.Render(payloadContent)

	// Labels if present
	labelsSection := ""
	if len(e.Labels) > 0 {
		var labelLines []string
		for k, v := range e.Labels {
			labelLines = append(labelLines, fmt.Sprintf("  %s: %s", k, v))
		}
		labelsSection = "\n" + lipgloss.NewStyle().
			Foreground(styles.ColorTextMuted).
			Render("Labels:\n"+strings.Join(labelLines, "\n"))
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		breadcrumb,
		"",
		metaCard,
		"",
		"Message:",
		payloadSection,
		labelsSection,
	)
}
