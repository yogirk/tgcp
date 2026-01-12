package logging

import (
	"strings"

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

	// Default: List View
	return s.renderListView()
}

// renderListView renders the main log table
func (s *Service) renderListView() string {
	doc := strings.Builder{}

	// Breadcrumb
	var crumbItems []string
	if s.heading != "" {
		crumbItems = []string{"Cloud Logging", s.heading}
	} else {
		crumbItems = []string{"Cloud Logging", "All Logs"}
	}

	// Debug: Show active filter in breadcrumb if set via text input
	if filterVal := s.filter.TextInput.Value(); filterVal != "" {
		// Truncate if too long (e.g. standard gce filter)
		displayFilter := filterVal
		if len(displayFilter) > 50 {
			displayFilter = "Filter: " + displayFilter[:47] + "..."
		} else {
			displayFilter = "Filter: " + displayFilter
		}
		crumbItems = append(crumbItems, displayFilter)
	}

	doc.WriteString(components.Breadcrumb(crumbItems...))
	doc.WriteString("\n\n")

	// Filter Bar
	// Always show filter bar for logging as it is crucial
	doc.WriteString(s.filter.View())
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

