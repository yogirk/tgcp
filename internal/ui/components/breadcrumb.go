package components

import (
	"strings"

	"github.com/yogirk/tgcp/internal/styles"
)

// Breadcrumb renders a consistent breadcrumb line.
// Empty segments are ignored.
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
	return styles.SubtleStyle.Render(strings.Join(segments, " > "))
}
