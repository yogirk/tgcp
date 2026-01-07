package logging

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/api/logging/v2"
	"google.golang.org/api/option"
)

// Client wraps the Logging API
type Client struct {
	svc       *logging.Service
	projectID string
}

// NewClient creates a new Logging client
func NewClient(ctx context.Context, projectID string) (*Client, error) {
	svc, err := logging.NewService(ctx, option.WithScopes(logging.LoggingReadScope))
	if err != nil {
		return nil, err
	}
	return &Client{
		svc:       svc,
		projectID: projectID,
	}, nil
}

// ListEntries fetches recent log entries
// filter: Advanced logs filter string (optional)
func (c *Client) ListEntries(ctx context.Context, filter string) ([]LogEntry, error) {
	req := &logging.ListLogEntriesRequest{
		ResourceNames: []string{"projects/" + c.projectID},
		OrderBy:       "timestamp desc",
		PageSize:      50,
	}

	if filter != "" {
		req.Filter = filter
	}

	resp, err := c.svc.Entries.List(req).Do()
	if err != nil {
		return nil, err
	}

	entries := make([]LogEntry, 0)
	for _, e := range resp.Entries {
		// Parse Timestamp
		ts, _ := time.Parse(time.RFC3339, e.Timestamp)

		// Parse Payload
		var payload string
		if e.TextPayload != "" {
			payload = e.TextPayload
		} else if e.JsonPayload != nil {
			// Quick stringify for table view
			payload = "{JSON Payload}"
		} else if e.ProtoPayload != nil {
			payload = "{Proto Payload}"
		}

		// Parse Resource
		var resType, resID string
		if e.Resource != nil {
			resType = e.Resource.Type
			if e.Resource.Labels != nil {
				// Try common labels
				if v, ok := e.Resource.Labels["instance_id"]; ok {
					resID = v
				} else if v, ok := e.Resource.Labels["function_name"]; ok {
					resID = v
				} else if v, ok := e.Resource.Labels["database_id"]; ok {
					resID = v
				} else if v, ok := e.Resource.Labels["service_name"]; ok {
					resID = v
				}
			}
		}

		// Full Payload (for details)
		fullPayload := payload
		if e.JsonPayload != nil {
			// Re-marshal map for pretty print? For now just raw map string
			fullPayload = fmt.Sprintf("%v", e.JsonPayload)
		}

		entries = append(entries, LogEntry{
			Timestamp:   ts,
			Severity:    e.Severity,
			Payload:     payload,
			Resource:    resType,
			ResourceID:  resID,
			InsertID:    e.InsertId,
			FullPayload: fullPayload,
		})
	}

	return entries, nil
}
