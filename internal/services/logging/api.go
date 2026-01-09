package logging

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"google.golang.org/api/logging/v2"
	"google.golang.org/api/option"
)

// Regex patterns for log cleaning
var (
	// Matches RFC3339-like timestamps at start of line
	// e.g. 2026-01-08T13:51:34.702339+00:00
	reTimestamp = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}[T\s]\d{2}:\d{2}:\d{2}(\.\d+)?(Z|[+-]\d{2}:?\d{2})?\s+`)

	// Matches Syslog headers
	// e.g. aryaka-kubeadm kubelet[1213]:
	reSyslogHeader = regexp.MustCompile(`^(\S+\s+)?\S+\[\d+\]:\s+`)

	// Matches standard Go/K8s log prefixes
	// e.g. I0108 13:51:34.701761    1213 scope.go:117]
	reK8sHeader = regexp.MustCompile(`^[IVWE]\d{4}\s+\d{2}:\d{2}:\d{2}\.\d+\s+\d+\s+\S+:\d+\]\s+`)
)

// Client wraps the Cloud Logging API (v2 REST)
type Client struct {
    service *logging.Service
    projectID string
}

// NewClient initializes a new Logging client using the v2 REST API
func NewClient(ctx context.Context, projectID string) (*Client, error) {
    // Use ADC with Logging Read scope
    svc, err := logging.NewService(ctx, option.WithScopes(logging.LoggingReadScope))
    if err != nil {
        return nil, fmt.Errorf("failed to create logging service: %w", err)
    }

    return &Client{
        service:   svc,
        projectID: projectID,
    }, nil
}

// Close is a no-op for the REST service wrapper
func (c *Client) Close() error {
    return nil
}

// ListEntries fetches log entries with pagination
func (c *Client) ListEntries(
    ctx context.Context,
    filter string,
    pageSize int,
    pageToken string,
) ([]LogEntry, string, error) {

    finalFilter := filter
    if finalFilter == "" {
        finalFilter = fmt.Sprintf(
            "timestamp >= \"%s\"",
            time.Now().Add(-30*time.Minute).Format(time.RFC3339),
        )
    }

    // Prepare request
    req := c.service.Entries.List(&logging.ListLogEntriesRequest{
        ResourceNames: []string{"projects/" + c.projectID},
        Filter:        finalFilter,
        PageSize:      int64(pageSize),
        PageToken:     pageToken,
        OrderBy:       "timestamp desc", // Equivalent to NewestFirst
    })

	resp, err := req.Context(ctx).Do()
	if err != nil {
		return nil, "", fmt.Errorf("failed to list entries: %w", err)
	}

	var entries []LogEntry
    for _, entry := range resp.Entries {
        // Parse Timestamp
        ts, _ := time.Parse(time.RFC3339Nano, entry.Timestamp)
        // Try fallback if Nano fails
        if ts.IsZero() {
            ts, _ = time.Parse(time.RFC3339, entry.Timestamp)
        }

        // Determine Payload and Severity
        payload := ""
        severity := strings.ToUpper(entry.Severity)

        if entry.TextPayload != "" {
            payload = cleanPayload(entry.TextPayload, ts)
        } else if len(entry.JsonPayload) > 0 {
			var data map[string]interface{}
			if err := json.Unmarshal(entry.JsonPayload, &data); err == nil {
				// Extract Severity from JSON if missing
				if severity == "" {
					if v, ok := data["severity"].(string); ok {
						severity = strings.ToUpper(v)
					}
				}

				// extract useful message
				if msg, ok := data["message"].(string); ok {
					payload = cleanPayload(msg, ts)
				} else if msg, ok := data["msg"].(string); ok {
					payload = cleanPayload(msg, ts)
				} else if msg, ok := data["log"].(string); ok {
					payload = cleanPayload(msg, ts)
				} else {
					// Fallback to raw JSON string
					payload = string(entry.JsonPayload)
				}
			} else {
				// Fallback to string
				payload = string(entry.JsonPayload)
			}
        } else if len(entry.ProtoPayload) > 0 {
            b, _ := json.Marshal(entry.ProtoPayload)
            payload = string(b)
        }

        // Extract Resource Info
        var (
            resourceType string
            resourceName string
            location     string
            projID       string
        )

        if entry.Resource != nil {
            resourceType = entry.Resource.Type
            if entry.Resource.Labels != nil {
                projID = entry.Resource.Labels["project_id"]
                location = entry.Resource.Labels["zone"]
                if location == "" {
                    location = entry.Resource.Labels["location"]
                }

                switch resourceType {
                case "gce_instance":
                    resourceName = entry.Resource.Labels["instance_id"]
                case "cloud_run_revision":
                    resourceName = entry.Resource.Labels["service_name"]
                case "k8s_container":
                    resourceName = fmt.Sprintf("%s/%s", entry.Resource.Labels["namespace_name"], entry.Resource.Labels["container_name"])
                default:
                    for _, v := range entry.Resource.Labels {
                        resourceName = v
                        break
                    }
                }
            }
        }

		if !isValidSeverity(severity) {
			severity = "DEFAULT"
		}

        entries = append(entries, LogEntry{
            Timestamp:    ts,
            Severity:     severity,
            Payload:      payload,

            ResourceType: resourceType,
            ResourceName: resourceName,
            Location:     location,
            ProjectID:    projID,

            LogName:      entry.LogName,
            Labels:       entry.Labels,
            InsertID:     entry.InsertId,
            FullPayload:  payload, // Simplified for now
        })
    }

    return entries, resp.NextPageToken, nil
}



// cleanPayload removes redundant timestamps and prefixes using regex
func cleanPayload(raw string, ts time.Time) string {
	// 1. Remove Timestamp
	// If the line starts with a timestamp string that looks like our log timestamp, strip it.
	// We rely on Regex for general shape match.
	if loc := reTimestamp.FindStringIndex(raw); loc != nil {
		raw = raw[loc[1]:]
	}

	// 2. Remove Syslog Header
	// e.g. "host app[123]: "
	if loc := reSyslogHeader.FindStringIndex(raw); loc != nil {
		raw = raw[loc[1]:]
	}
	
	// 3. Remove K8s Header
	// e.g. "I0108 ... ] "
	if loc := reK8sHeader.FindStringIndex(raw); loc != nil {
		raw = raw[loc[1]:]
	}

	// 4. Remove quote wrapper if present (common in "msg" fields)
	if strings.HasPrefix(raw, "\"") && strings.HasSuffix(raw, "\"") {
		raw = strings.Trim(raw, "\"")
	}

	return strings.TrimSpace(raw)
}

func isValidSeverity(s string) bool {
	switch s {
	case "DEFAULT", "DEBUG", "INFO", "NOTICE", "WARNING", "ERROR", "CRITICAL", "ALERT", "EMERGENCY":
		return true
	}
	return false
}
