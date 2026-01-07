package logging

import "time"

// LogEntry represents a unified log entry for display
type LogEntry struct {
	Timestamp   time.Time
	Severity    string
	Payload     string
	Resource    string // e.g. "gce_instance"
	ResourceID  string // e.g. "my-vm"
	InsertID    string
	FullPayload string // For detail view
}
