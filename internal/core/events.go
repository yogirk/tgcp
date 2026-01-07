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
type SwitchToLogsMsg struct {
	Filter string
}
