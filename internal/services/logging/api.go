package logging

import (
    "context"
    "encoding/json"
    "fmt"
    "strings"
    "time"

    "google.golang.org/api/logging/v2"
    "google.golang.org/api/option"
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

        // Determine Payload
        payload := ""
        if entry.TextPayload != "" {
            payload = entry.TextPayload
        } else if len(entry.JsonPayload) > 0 {
            // JsonPayload is raw JSON map
            b, _ := json.Marshal(entry.JsonPayload)
            payload = string(b)
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

        // Severity
        severity := strings.ToUpper(entry.Severity)

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


