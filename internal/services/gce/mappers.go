package gce

import (
	"github.com/charmbracelet/bubbles/table"
)

// InstanceToRow converts an Instance model to a table.Row
func InstanceToRow(i Instance) table.Row {
	var statusStr string
	switch i.State {
	case StateRunning:
		statusStr = "🟢 " + string(i.State)
	case StateStopped:
		statusStr = "🔴 " + string(i.State)
	case StateTerminated:
		statusStr = "🔴 STOP"
	case StateProvisioning, StateStaging, StateStopping, StateSuspending, StateRepairing:
		statusStr = "🔄 " + string(i.State)
	default:
		statusStr = "⚪ " + string(i.State)
	}

	// Calculate Total Disk Size (Optional: kept calculation if needed, but removing from row)
	/*
		var totalDisk int64
		for _, d := range i.Disks {
			totalDisk += d.SizeGB
		}
		diskStr := fmt.Sprintf("%dGB", totalDisk)
	*/

	// Column Order: Name, Status, Zone, Internal IP, External IP, ID
	return table.Row{
		i.Name,
		statusStr,
		i.Zone,
		i.InternalIP,
		i.ExternalIP,
		i.ID,
	}
}

// GetGCEColumns returns the table column definitions
func GetGCEColumns() []table.Column {
	// Simple responsive logic can go here
	return []table.Column{
		{Title: "VM Name", Width: 30},
		{Title: "VM STATE", Width: 20}, // Increased width to prevent dot-only issue
		{Title: "GCP Zone", Width: 15},
		{Title: "Int. IP", Width: 15},
		{Title: "Ext. IP", Width: 15},
		{Title: "ID", Width: 20}, // Hidden or at end
	}
}
