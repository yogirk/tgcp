package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/yogirk/tgcp/internal/styles"
)

// Breadcrumb separator
const breadcrumbSep = " › "

// Breadcrumb renders a consistent breadcrumb line.
// Empty segments are ignored.
// Path segments are muted, current location (last segment) is prominent.
func Breadcrumb(parts ...string) string {
	segments := make([]string, 0, len(parts))
	for _, part := range parts {
		if strings.TrimSpace(part) == "" {
			continue
		}
		segments = append(segments, part)
	}
	if len(segments) == 0 {
		return ""
	}

	// Style definitions
	pathStyle := styles.SubtleStyle
	sepStyle := lipgloss.NewStyle().Foreground(styles.ColorBorderSubtle)
	currentStyle := lipgloss.NewStyle().
		Foreground(styles.ColorTextPrimary).
		Bold(true)

	// Single segment - just render it as current location
	if len(segments) == 1 {
		return currentStyle.Render(segments[0])
	}

	// Multiple segments: path (muted) + separator + current (prominent)
	pathParts := segments[:len(segments)-1]
	current := segments[len(segments)-1]

	// Render path segments with separators
	var renderedPath []string
	for i, part := range pathParts {
		renderedPath = append(renderedPath, pathStyle.Render(part))
		if i < len(pathParts)-1 {
			renderedPath = append(renderedPath, sepStyle.Render(breadcrumbSep))
		}
	}

	path := strings.Join(renderedPath, "")
	sep := sepStyle.Render(breadcrumbSep)
	curr := currentStyle.Render(current)

	return path + sep + curr
}
