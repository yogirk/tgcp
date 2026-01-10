package gce

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/yogirk/tgcp/internal/core"
	compute "google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

// Client wraps the GCE API service
type Client struct {
	service *compute.Service
}

// NewClient initializes a new GCE API client
func NewClient(ctx context.Context) (*Client, error) {
	httpClient, err := core.NewHTTPClient(ctx, compute.ComputeScope)
	if err != nil {
		return nil, fmt.Errorf("failed to create http client: %w", err)
	}

	svc, err := compute.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("failed to create compute service: %w", err)
	}
	return &Client{service: svc}, nil
}

// ListInstances fetches all instances across all zones (AggregatedList)
func (c *Client) ListInstances(projectID string) ([]Instance, error) {
	req := c.service.Instances.AggregatedList(projectID)
	var instances []Instance

	if err := req.Pages(context.Background(), func(page *compute.InstanceAggregatedList) error {
		for zoneKey, items := range page.Items {
			// zoneKey format: "zones/us-central1-a"
			zone := strings.TrimPrefix(zoneKey, "zones/")

			// Skip scopes that might be regions or warnings
			if len(items.Instances) == 0 {
				continue
			}

			for _, inst := range items.Instances {
				// Parse Network Interfaces
				var internalIP, externalIP string
				if len(inst.NetworkInterfaces) > 0 {
					internalIP = inst.NetworkInterfaces[0].NetworkIP
					if len(inst.NetworkInterfaces[0].AccessConfigs) > 0 {
						externalIP = inst.NetworkInterfaces[0].AccessConfigs[0].NatIP
					}
				}

				// Parse Machine Type
				// Format: "https://www.googleapis.com/compute/v1/projects/proj/zones/zone/machineTypes/n1-standard-1"
				parts := strings.Split(inst.MachineType, "/")
				machineType := parts[len(parts)-1]

				// Parse Disks
				var disks []Disk
				for _, d := range inst.Disks {
					// d.Type is not always populated or is a URL
					// If boot disk, it might be in InitializeParams but AttachedDisk also has DiskSizeGb
					diskType := "pd-standard" // Default
					// Try to guess from InitializeParams if exists
					if d.InitializeParams != nil && d.InitializeParams.DiskType != "" {
						// Format: zones/.../diskTypes/pd-ssd
						dtParts := strings.Split(d.InitializeParams.DiskType, "/")
						diskType = dtParts[len(dtParts)-1]
					}

					disks = append(disks, Disk{
						Name:   d.DeviceName,
						SizeGB: d.DiskSizeGb,
						Type:   diskType,
					})
				}

				// Parse Creation Time
				creationTime, _ := time.Parse(time.RFC3339, inst.CreationTimestamp)

				// Determine OS Image
				osImage := "Unknown"
				for _, d := range inst.Disks {
					if d.Boot {
						// Try InitializeParams first
						if d.InitializeParams != nil && d.InitializeParams.SourceImage != "" {
							parts := strings.Split(d.InitializeParams.SourceImage, "/")
							osImage = parts[len(parts)-1]
						} else if len(d.Licenses) > 0 {
							// Fallback to licenses
							parts := strings.Split(d.Licenses[0], "/")
							osImage = parts[len(parts)-1]
						}
						break
					}
				}

				instances = append(instances, Instance{
					ID:           fmt.Sprintf("%d", inst.Id),
					Name:         inst.Name,
					Zone:         zone,
					State:        InstanceState(inst.Status), // Simplified cast
					MachineType:  machineType,
					InternalIP:   internalIP,
					ExternalIP:   externalIP,
					CreationTime: creationTime,
					Tags:         inst.Tags.Items,
					Disks:        disks,
					OSImage:      osImage,
				})
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return instances, nil
}

// StartInstance starts a stopped instance
func (c *Client) StartInstance(projectID, zone, instanceName string) error {
	_, err := c.service.Instances.Start(projectID, zone, instanceName).Do()
	return err
}

// StopInstance stops a running instance
func (c *Client) StopInstance(projectID, zone, instanceName string) error {
	_, err := c.service.Instances.Stop(projectID, zone, instanceName).Do()
	return err
}
