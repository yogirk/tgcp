package components

import (
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
