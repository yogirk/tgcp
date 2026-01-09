package core

import "time"

// StatusMsg updates the status bar message
type StatusMsg struct {
	Message string
	IsError bool
}

// LastUpdatedMsg updates the "Last Updated" timestamp
type LastUpdatedMsg time.Time

// SwitchToLogsMsg requests a context switch to the logging service
// SwitchToLogsMsg requests a context switch to the logging service
type SwitchToLogsMsg struct {
	Filter string
	Source string // The short name of the service initiating the switch
	Heading string // Optional heading to display (e.g. resource name)
}

// SwitchToServiceMsg requests a context switch to a specific service
type SwitchToServiceMsg struct {
	Service string // The short name of the service to switch to
}
