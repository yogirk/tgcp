package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/yogirk/tgcp/internal/styles"
)

// KeyValue represents a single key/value row in a detail card.
type KeyValue struct {
	Key           string
	Value         string
	UseValueStyle bool
}

// DetailCardOpts controls the look and content of a detail card.
type DetailCardOpts struct {
	Title       string
	Rows        []KeyValue
	Width       int
	BorderColor lipgloss.Color
	FooterHint  string
}

// DetailCard renders a standardized detail card with header bar.
func DetailCard(opts DetailCardOpts) string {
	width := opts.Width
	if width <= 0 {
		width = 80
	}
	borderColor := opts.BorderColor
	if borderColor == "" {
		borderColor = styles.ColorBorderSubtle
	}

	// Header bar style (matches table headers)
	title := styles.HeaderStyle.Copy().
		Width(width).
		Render(opts.Title)

	body := renderKeyValues(opts.Rows, styles.LabelStyle, styles.ValueStyle)

	box := styles.PrimaryBoxStyle.Copy().
		BorderForeground(borderColor).
		Width(width).
		Render(body)

	parts := []string{title, box}
	if opts.FooterHint != "" {
		parts = append(parts, RenderFooterHint(opts.FooterHint))
	}

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// DetailSection renders a secondary section box for detail views.
func DetailSection(title, body string, borderColor lipgloss.Color) string {
	if borderColor == "" {
		borderColor = styles.ColorBorderSubtle
	}
	content := lipgloss.JoinVertical(lipgloss.Left, styles.HeaderStyle.Render(title), body)
	return styles.SecondaryBoxStyle.Copy().
		BorderForeground(borderColor).
		Width(80).
		Render(content)
}

// statusKeys are field names that should be auto-rendered as status badges
var statusKeys = map[string]bool{
	"Status": true,
	"State":  true,
}

func renderKeyValues(rows []KeyValue, keyStyle, valueStyle lipgloss.Style) string {
	if len(rows) == 0 {
		return ""
	}
	maxKeyLen := 0
	for _, row := range rows {
		if len(row.Key) > maxKeyLen {
			maxKeyLen = len(row.Key)
		}
	}

	lines := make([]string, 0, len(rows))
	for _, row := range rows {
		key := fmt.Sprintf("%-*s", maxKeyLen+1, row.Key+":")
		value := row.Value

		// Auto-render status fields if not already styled
		if statusKeys[row.Key] && !strings.Contains(row.Value, "\x1b") {
			value = RenderStatus(row.Value)
		} else if row.UseValueStyle || !strings.Contains(row.Value, "\x1b") {
			value = valueStyle.Render(row.Value)
		}

		line := fmt.Sprintf("%s %s", keyStyle.Render(key), value)
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

// RenderFooterHint parses a hint string like "s Start | x Stop | q Back"
// and renders it with highlighted keys: [s] Start  [x] Stop  [q] Back
// This function is exported so it can be used by any component that needs
// styled keyboard hints (not just DetailCard).
func RenderFooterHint(hint string) string {
	// Style for the key badge [s]
	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("232")).
		Background(styles.ColorBorderSubtle).
		Bold(true).
		Padding(0, 0)

	// Style for the action text
	actionStyle := lipgloss.NewStyle().
		Foreground(styles.ColorTextMuted)

	// Split by |
	parts := strings.Split(hint, "|")
	var rendered []string

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Split into key and action (first word is key, rest is action)
		words := strings.SplitN(part, " ", 2)
		if len(words) == 0 {
			continue
		}

		key := words[0]
		action := ""
		if len(words) > 1 {
			action = words[1]
		}

		// Render as [key] Action
		keyBadge := keyStyle.Render("[" + key + "]")
		actionText := actionStyle.Render(action)

		rendered = append(rendered, keyBadge+" "+actionText)
	}

	return strings.Join(rendered, "  ")
}
