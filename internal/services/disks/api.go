package disks

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/api/compute/v1"
)

type Client struct {
	service *compute.Service
}

func NewClient(ctx context.Context) (*Client, error) {
	svc, err := compute.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("compute client: %w", err)
	}
	return &Client{service: svc}, nil
}

func (c *Client) ListDisks(projectID string) ([]Disk, error) {
	// Use aggregated list to get disks from all zones
	req := c.service.Disks.AggregatedList(projectID)
	var disks []Disk

	if err := req.Pages(context.Background(), func(page *compute.DiskAggregatedList) error {
		for _, scopedList := range page.Items {
			for _, d := range scopedList.Disks {
				// Parse Zone from URL: https://www.googleapis.com/compute/v1/projects/.../zones/us-central1-a
				zone := ""
				parts := strings.Split(d.Zone, "/")
				if len(parts) > 0 {
					zone = parts[len(parts)-1]
				}

				disks = append(disks, Disk{
					Name:                d.Name,
					Zone:                zone,
					SizeGb:              d.SizeGb,
					Type:                d.Type,
					Status:              d.Status,
					LastAttachTimestamp: d.LastAttachTimestamp,
					Users:               d.Users,
					SourceImage:         d.SourceImage,
				})
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return disks, nil
}
