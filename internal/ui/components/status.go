package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/yogirk/tgcp/internal/styles"
)

// StatusCategory represents the type of status for styling purposes
type StatusCategory int

const (
	StatusRunning StatusCategory = iota
	StatusStopped
	StatusPending
	StatusUnknown
)

// Status icons
const (
	IconRunning = "✓"
	IconStopped = "✗"
	IconPending = "◐"
	IconUnknown = "○"
)

// statusConfig holds the styling configuration for each status category
type statusConfig struct {
	icon       string
	foreground lipgloss.Color
	background lipgloss.Color
}

var statusConfigs = map[StatusCategory]statusConfig{
	StatusRunning: {
		icon:       IconRunning,
		foreground: lipgloss.Color("0"),   // Black text for contrast
		background: lipgloss.Color("42"),  // Green background
	},
	StatusStopped: {
		icon:       IconStopped,
		foreground: lipgloss.Color("0"),   // Black text for contrast
		background: lipgloss.Color("196"), // Red background
	},
	StatusPending: {
		icon:       IconPending,
		foreground: lipgloss.Color("0"),   // Black text for contrast
		background: lipgloss.Color("214"), // Yellow/Orange background
	},
	StatusUnknown: {
		icon:       IconUnknown,
		foreground: lipgloss.Color("252"), // Light text
		background: lipgloss.Color("240"), // Grey background
	},
}

// runningStates maps state strings to StatusRunning
var runningStates = map[string]bool{
	"RUNNING":   true,
	"READY":     true,
	"ACTIVE":    true,
	"DONE":      true,
	"RUNNABLE":  true,
	"SUCCEEDED": true,
	"HEALTHY":   true,
	"ENABLED":   true,
	"NOTICE":    true,
	"INFO":      true,
}

// stoppedStates maps state strings to StatusStopped
var stoppedStates = map[string]bool{
	"STOPPED":    true,
	"TERMINATED": true,
	"FAILED":     true,
	"ERROR":      true,
	"DELETED":    true,
	"SUSPENDED":  true,
	"OFFLINE":    true,
	"DISABLED":   true,
	"CANCELLED":  true,
}

// pendingStates maps state strings to StatusPending
var pendingStates = map[string]bool{
	"PENDING":          true,
	"PROVISIONING":     true,
	"STAGING":          true,
	"STOPPING":         true,
	"SUSPENDING":       true,
	"REPAIRING":        true,
	"STARTING":         true,
	"UPDATING":         true,
	"CREATING":         true,
	"DELETING":         true,
	"MAINTENANCE":      true,
	"RECONCILING":      true,
	"JOB_STATE_QUEUED": true,
	"DRAINING":         true,
	"CANCELLING":       true,
	"WARNING":          true,
}

// CategorizeStatus determines the StatusCategory for a given state string
func CategorizeStatus(state string) StatusCategory {
	upper := strings.ToUpper(strings.TrimSpace(state))

	if runningStates[upper] {
		return StatusRunning
	}
	if stoppedStates[upper] {
		return StatusStopped
	}
	if pendingStates[upper] {
		return StatusPending
	}
	return StatusUnknown
}

// RenderStatus renders a status badge with icon and background color
// Example output: " ✓ RUNNING " with green background
func RenderStatus(state string) string {
	category := CategorizeStatus(state)
	config := statusConfigs[category]

	// Clean up the display text
	displayText := strings.ToUpper(strings.TrimSpace(state))

	// Shorten some verbose states for display
	displayText = shortenState(displayText)

	badge := lipgloss.NewStyle().
		Foreground(config.foreground).
		Background(config.background).
		Padding(0, 1).
		Render(config.icon + " " + displayText)

	return badge
}

// RenderStatusMinimal renders just the icon with color (no background)
// Useful for tight spaces like table cells
func RenderStatusMinimal(state string) string {
	category := CategorizeStatus(state)
	config := statusConfigs[category]

	displayText := strings.ToUpper(strings.TrimSpace(state))
	displayText = shortenState(displayText)

	// Use the background color as foreground for the icon (it's more vibrant)
	icon := lipgloss.NewStyle().
		Foreground(config.background).
		Bold(true).
		Render(config.icon)

	text := lipgloss.NewStyle().
		Foreground(styles.ColorTextPrimary).
		Render(" " + displayText)

	return icon + text
}

// StatusSummary renders a one-line breakdown of item statuses as inline
// pills. Designed to sit above a list view so you get an at-a-glance sense
// of the set: "✓ 4 Running  ·  ✗ 1 Stopped  ·  5 total".
//
// Only categories with a non-zero count appear. The total is always shown.
func StatusSummary(states []string) string {
	if len(states) == 0 {
		return ""
	}

	counts := map[StatusCategory]int{}
	for _, s := range states {
		counts[CategorizeStatus(s)]++
	}

	order := []struct {
		cat   StatusCategory
		label string
	}{
		{StatusRunning, "Running"},
		{StatusPending, "Pending"},
		{StatusStopped, "Stopped"},
		{StatusUnknown, "Unknown"},
	}

	var parts []string
	for _, item := range order {
		n := counts[item.cat]
		if n == 0 {
			continue
		}
		parts = append(parts, renderCountPill(item.cat, n, item.label))
	}

	total := styles.SubtleStyle.Render(fmt.Sprintf("%d total", len(states)))
	parts = append(parts, total)

	sep := styles.SubtleStyle.Render("  ·  ")
	return strings.Join(parts, sep)
}

// renderCountPill formats a single "icon N label" segment with the status
// category's accent colour.
func renderCountPill(cat StatusCategory, count int, label string) string {
	cfg := statusConfigs[cat]
	icon := lipgloss.NewStyle().Foreground(cfg.background).Bold(true).Render(cfg.icon)
	num := lipgloss.NewStyle().Foreground(styles.ColorTextPrimary).Bold(true).Render(fmt.Sprintf("%d", count))
	lbl := lipgloss.NewStyle().Foreground(styles.ColorTextMuted).Render(label)
	return icon + " " + num + " " + lbl
}

// shortenState shortens verbose state names for display
func shortenState(state string) string {
	switch state {
	case "TERMINATED":
		return "STOPPED"
	case "JOB_STATE_QUEUED":
		return "QUEUED"
	case "JOB_STATE_RUNNING":
		return "RUNNING"
	case "JOB_STATE_DONE":
		return "DONE"
	case "JOB_STATE_FAILED":
		return "FAILED"
	case "JOB_STATE_CANCELLED":
		return "CANCELLED"
	default:
		return state
	}
}
