package gke

import (
	"context"
	"fmt"

	"google.golang.org/api/container/v1"
	"google.golang.org/api/option"
)

type Client struct {
	service *container.Service
}

func NewClient(ctx context.Context) (*Client, error) {
	svc, err := container.NewService(ctx, option.WithScopes(container.CloudPlatformScope))
	if err != nil {
		return nil, fmt.Errorf("gke client: %w", err)
	}
	return &Client{service: svc}, nil
}

func (c *Client) ListClusters(projectID string) ([]Cluster, error) {
	// Use aggregated list to get clusters from all zones/regions
	parent := fmt.Sprintf("projects/%s/locations/-", projectID)
	resp, err := c.service.Projects.Locations.Clusters.List(parent).Do()
	if err != nil {
		return nil, err
	}

	var clusters []Cluster
	for _, cl := range resp.Clusters {
		// Parse location from selflink or name if needed,
		// but List response struct usually has Location field if we used parent with location "-"
		// Actually for aggregated list we usually use projects/{projectId}/locations/-
		// Let's verify if the above List call supports "-"

		clusters = append(clusters, Cluster{
			Name:          cl.Name,
			Location:      cl.Location,
			Status:        cl.Status,
			MasterVersion: cl.CurrentMasterVersion,
			Endpoint:      cl.Endpoint,
			Network:       cl.Network,
			Subnetwork:    cl.Subnetwork,
			NodeCount:     int(cl.CurrentNodeCount),
			Mode:          getMode(cl),
			SelfLink:      cl.SelfLink,
			NodePools:     convertNodePools(cl.NodePools),
		})
	}
	return clusters, nil
}

// Helpers

func getMode(cl *container.Cluster) string {
	if cl.Autopilot != nil && cl.Autopilot.Enabled {
		return "Autopilot"
	}
	return "Standard"
}

func convertNodePools(apiPools []*container.NodePool) []NodePool {
	var pools []NodePool
	for _, p := range apiPools {
		var machineType string
		var diskSize int64
		var isSpot bool
		var minCount, maxCount int64

		if p.Config != nil {
			machineType = p.Config.MachineType
			diskSize = p.Config.DiskSizeGb
			isSpot = p.Config.Spot
		}

		if p.Autoscaling != nil {
			minCount = p.Autoscaling.MinNodeCount
			maxCount = p.Autoscaling.MaxNodeCount
		}

		pools = append(pools, NodePool{
			Name:             p.Name,
			Status:           p.Status,
			MachineType:      machineType,
			DiskSizeGb:       diskSize,
			InitialNodeCount: p.InitialNodeCount,
			Autoscaling: AutoscalingConfig{
				Enabled:      p.Autoscaling != nil && p.Autoscaling.Enabled,
				MinNodeCount: minCount,
				MaxNodeCount: maxCount,
			},
			IsSpot:  isSpot,
			Version: p.Version,
		})
	}
	return pools
}
