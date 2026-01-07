package gce

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/rk/tgcp/internal/styles"
)

// InstanceToRow converts an Instance model to a table.Row
func InstanceToRow(i Instance) table.Row {
	statusStyle := lipgloss.NewStyle()
	switch i.State {
	case StateRunning:
		statusStyle = statusStyle.Foreground(styles.ColorSuccess) // Green
	case StateStopped, StateTerminated:
		statusStyle = statusStyle.Foreground(styles.ColorSubtext) // Grey/Dim
	default:
		statusStyle = statusStyle.Foreground(styles.ColorWarning) // Orange
	}

	return table.Row{
		i.Name,
		string(i.State), // Remove style for debug
		i.Zone,
		i.InternalIP,
		i.ExternalIP,
		i.ID,
	}
}

// BuildColumns returns the table column definitions
func BuildColumns(width int) []table.Column {
	// Simple responsive logic can go here
	return []table.Column{
		{Title: "ID", Width: 20}, // Hidden mostly
		{Title: "Name", Width: 30},
		{Title: "Zone", Width: 15},
		{Title: "Status", Width: 12},
		{Title: "Internal IP", Width: 15},
		{Title: "External IP", Width: 15},
	}
}
