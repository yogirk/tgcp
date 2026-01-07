package gce

import (
	"context"
	"fmt"
	"strings"

	"github.com/rk/tgcp/internal/core"
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

				instances = append(instances, Instance{
					ID:          fmt.Sprintf("%d", inst.Id),
					Name:        inst.Name,
					Zone:        zone,
					State:       InstanceState(inst.Status), // Simplified cast
					MachineType: machineType,
					InternalIP:  internalIP,
					ExternalIP:  externalIP,
					// Tags: inst.Tags.Items, // Check if Tags struct exists
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
