package bigtable

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/api/bigtableadmin/v2"
)

type Client struct {
	service *bigtableadmin.Service
}

func NewClient(ctx context.Context) (*Client, error) {
	svc, err := bigtableadmin.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("bigtable client: %w", err)
	}
	return &Client{service: svc}, nil
}

func (c *Client) ListInstances(projectID string) ([]Instance, error) {
	var instances []Instance
	parent := fmt.Sprintf("projects/%s", projectID)

	// Note: Bigtable API handles pagination, but for MVP we take the first page or iterate
	// We'll trust the default page size is sufficient or loop
	// Using Pages() helper is safest

	call := c.service.Projects.Instances.List(parent)
	err := call.Pages(context.Background(), func(page *bigtableadmin.ListInstancesResponse) error {
		for _, i := range page.Instances {
			parts := strings.Split(i.Name, "/")
			shortName := parts[len(parts)-1]

			// Enum mapping for State/Type if needed, but strings are usually fine

			instances = append(instances, Instance{
				Name:        shortName,
				DisplayName: i.DisplayName,
				ProjectID:   projectID,
				State:       i.State,
				Type:        i.Type,
			})
		}
		return nil
	})
	return instances, err
}

func (c *Client) ListClusters(projectID, instanceID string) ([]Cluster, error) {
	var clusters []Cluster
	parent := fmt.Sprintf("projects/%s/instances/%s", projectID, instanceID)

	call := c.service.Projects.Instances.Clusters.List(parent)
	err := call.Pages(context.Background(), func(page *bigtableadmin.ListClustersResponse) error {
		for _, cl := range page.Clusters {
			parts := strings.Split(cl.Name, "/")
			shortName := parts[len(parts)-1]

			zoneParts := strings.Split(cl.Location, "/")
			zone := ""
			if len(zoneParts) > 0 {
				zone = zoneParts[len(zoneParts)-1]
			}

			clusters = append(clusters, Cluster{
				Name:        shortName,
				Zone:        zone,
				ServeNodes:  int(cl.ServeNodes),
				State:       cl.State,
				StorageType: cl.DefaultStorageType,
			})
		}
		return nil
	})
	return clusters, err
}
