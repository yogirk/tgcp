package dataproc

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/api/dataproc/v1"
)

type Client struct {
	service *dataproc.Service
}

func NewClient(ctx context.Context) (*Client, error) {
	svc, err := dataproc.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("dataproc client: %w", err)
	}
	return &Client{service: svc}, nil
}

func (c *Client) ListClusters(projectID string, region string) ([]Cluster, error) {
	var clusters []Cluster

	err := c.service.Projects.Regions.Clusters.List(projectID, region).Pages(context.Background(), func(page *dataproc.ListClustersResponse) error {
		for _, cl := range page.Clusters {
			status := "UNKNOWN"
			if cl.Status != nil {
				status = cl.Status.State
			}

			masterType := "N/A"
			workerType := "N/A"
			workerCount := 0
			zone := ""

			if cl.Config != nil {
				if cl.Config.MasterConfig != nil {
					masterType = machineTypeShort(cl.Config.MasterConfig.MachineTypeUri)
				}
				if cl.Config.WorkerConfig != nil {
					workerType = machineTypeShort(cl.Config.WorkerConfig.MachineTypeUri)
					workerCount = int(cl.Config.WorkerConfig.NumInstances)
				}
				if cl.Config.GceClusterConfig != nil {
					zone = machineTypeShort(cl.Config.GceClusterConfig.ZoneUri) // Reuse shortener for zone
				}
			}

			clusters = append(clusters, Cluster{
				Name:          cl.ClusterName,
				ProjectID:     projectID,
				Status:        status,
				MasterMachine: masterType,
				WorkerCount:   workerCount,
				WorkerMachine: workerType,
				Zone:          zone,
			})
		}
		return nil
	})
	return clusters, err
}

// machineTypeShort extracts "n1-standard-4" from full URI
func machineTypeShort(uri string) string {
	// .../zones/us-central1-a/machineTypes/n1-standard-4
	// or .../zones/us-central1-a
	parts := parseURI(uri) // Simplified check
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return uri
}

func parseURI(uri string) []string {
	// Simple splitter
	return strings.Split(uri, "/")
}
