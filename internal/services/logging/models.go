package logging

import "time"

// LogEntry represents a unified log entry for display
type LogEntry struct {
    Timestamp   time.Time
    Severity    string
    Payload     string

    ResourceType string            // gce_instance, cloud_run_revision
    ResourceName string            // vm-name, service-name
    Location     string            // us-central1
    ProjectID   string

    LogName     string
    Labels      map[string]string

    InsertID    string
    FullPayload string
}

