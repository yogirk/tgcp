package logging

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/logging"
	"cloud.google.com/go/logging/logadmin"
	"google.golang.org/api/iterator"
)

// Client wraps the Cloud Logging API
type Client struct {
	client *logging.Client
	admin  *logadmin.Client
}

// NewClient initializes a new Logging client
func NewClient(ctx context.Context, projectID string) (*Client, error) {
	client, err := logging.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create logging client: %w", err)
	}

	admin, err := logadmin.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create logadmin client: %w", err)
	}

	return &Client{
		client: client,
		admin:  admin,
	}, nil
}

// Close closes the clients
func (c *Client) Close() error {
	if c.client != nil {
		c.client.Close()
	}
	if c.admin != nil {
		return c.admin.Close()
	}
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

	it := c.admin.Entries(
		ctx,
		logadmin.Filter(finalFilter),
		logadmin.NewestFirst(),
	)
	pager := iterator.NewPager(it, pageSize, pageToken)

	var rawEntries []*logging.Entry
	nextPageToken, err := pager.NextPage(&rawEntries)
	if err != nil {
		return nil, "", err
	}

	var entries []LogEntry

	for _, entry := range rawEntries {

		// -------- Payload --------
		payload := ""
		switch p := entry.Payload.(type) {
		case string:
			payload = p
		default:
			payload = fmt.Sprintf("%v", p)
		}

		// -------- Resource Extraction (STEP 2) --------
		var (
			resourceType string
			resourceName string
			location     string
			projectID    string
		)

		if entry.Resource != nil {
			res := entry.Resource
			resourceType = res.Type

			// Common labels across services
			projectID = res.Labels["project_id"]
			location = res.Labels["zone"]
			if location == "" {
				location = res.Labels["location"]
			}

			// Resource-specific name resolution
			switch res.Type {
			case "gce_instance":
				resourceName = res.Labels["instance_id"]

			case "cloud_run_revision":
				resourceName = res.Labels["service_name"]

			case "k8s_container":
				resourceName = fmt.Sprintf(
					"%s/%s",
					res.Labels["namespace_name"],
					res.Labels["container_name"],
				)

			default:
				// Fallback: pick any meaningful label
				for _, v := range res.Labels {
					resourceName = v
					break
				}
			}
		}

		entries = append(entries, LogEntry{
			Timestamp:    entry.Timestamp,
			Severity:     entry.Severity.String(),
			Payload:      payload,

			ResourceType: resourceType,
			ResourceName: resourceName,
			Location:     location,
			ProjectID:    projectID,

			LogName:      entry.LogName,
			Labels:       entry.Labels,
			InsertID:     entry.InsertID,
			FullPayload:  fmt.Sprintf("%+v", entry.Payload),
		})
	}

	return entries, nextPageToken, nil
}

