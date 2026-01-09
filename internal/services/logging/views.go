package logging

import (
	"fmt"
	"strings"

	"github.com/rk/tgcp/internal/styles"
)

// View renders the service UI
func (s *Service) View() string {
	if s.loading {
		return "Loading logs..."
	}
	if s.err != nil {
		return fmt.Sprintf("Error: %v", s.err)
	}

	// Default: List View
	return s.renderListView()
}

// renderListView renders the main log table
func (s *Service) renderListView() string {
	doc := strings.Builder{}

	// Heading
	if s.heading != "" {
		header := styles.TitleStyle.Copy().
			Foreground(styles.ColorPrimary).
			Bold(true).
			MarginBottom(1).
			Render(s.heading)
		doc.WriteString(header)
		doc.WriteString("\n")
	}

	// Filter Bar
	// Always show filter bar for logging as it is crucial
	if s.filtering {
		doc.WriteString(s.filterInput.View())
	} else {
		// Render as wrapped text for better visibility when not editing
		filterVal := s.filterInput.Value()
		if filterVal == "" {
			filterVal = s.filterInput.Placeholder
		} else {
			// Add prompt for consistency
			filterVal = "/ " + filterVal
		}
		// Use a style that mimics input but wraps
		style := styles.BaseStyle.Copy().
			Width(s.viewport.Width - 4). // Match viewport wrapping
			Foreground(styles.ColorSubtext)
		
		doc.WriteString(style.Render(filterVal))
	}
	doc.WriteString("\n")

	doc.WriteString(styles.BaseStyle.Render(s.viewport.View()))
	return doc.String()
}

func renderSeverity(severity string) string {
	switch severity {
	case "ERROR", "CRITICAL", "ALERT", "EMERGENCY":
		return styles.ErrorStyle.Render("🔴 " + severity)
	case "WARNING":
		return styles.WarningStyle.Render("🟡 " + severity)
	case "NOTICE", "INFO":
		return styles.SuccessStyle.Render("🔵 " + severity)
	default:
		return severity
	}
}
