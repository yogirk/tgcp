package core

import "time"

// StatusMsg updates the status bar message
type StatusMsg struct {
	Message string
	IsError bool
}

// LastUpdatedMsg updates the "Last Updated" timestamp
type LastUpdatedMsg time.Time
