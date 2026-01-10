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

// DetailCard renders a standardized ASCII-style detail card.
func DetailCard(opts DetailCardOpts) string {
	width := opts.Width
	if width <= 0 {
		width = 80
	}
	borderColor := opts.BorderColor
	if borderColor == "" {
		borderColor = styles.ColorBorderSubtle
	}

	title := styles.LabelStyle.Render(fmt.Sprintf("╭─ %s ─", opts.Title))
	body := renderKeyValues(opts.Rows, styles.LabelStyle, styles.ValueStyle)

	box := styles.BoxStyle.Copy().
		BorderForeground(borderColor).
		Padding(1).
		Width(width).
		Render(body)

	parts := []string{title, box}
	if opts.FooterHint != "" {
		parts = append(parts, styles.HelpStyle.Render(opts.FooterHint))
	}

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// DetailSection renders a secondary section box for detail views.
func DetailSection(title, body string, borderColor lipgloss.Color) string {
	if borderColor == "" {
		borderColor = styles.ColorBorderSubtle
	}
	content := lipgloss.JoinVertical(lipgloss.Left, styles.HeaderStyle.Render(title), body)
	return styles.BoxStyle.Copy().
		BorderForeground(borderColor).
		Width(80).
		Render(content)
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
		if row.UseValueStyle || !strings.Contains(row.Value, "\x1b") {
			value = valueStyle.Render(row.Value)
		}
		line := fmt.Sprintf("%s %s", keyStyle.Render(key), value)
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}
