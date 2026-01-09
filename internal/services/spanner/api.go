package spanner

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/api/spanner/v1"
)

type Client struct {
	service *spanner.Service
}

func NewClient(ctx context.Context) (*Client, error) {
	svc, err := spanner.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("spanner client: %w", err)
	}
	return &Client{service: svc}, nil
}

func (c *Client) ListInstances(projectID string) ([]Instance, error) {
	var instances []Instance
	parent := fmt.Sprintf("projects/%s", projectID)

	err := c.service.Projects.Instances.List(parent).Pages(context.Background(), func(page *spanner.ListInstancesResponse) error {
		for _, i := range page.Instances {
			// Name: projects/{project}/instances/{instance}
			parts := strings.Split(i.Name, "/")
			shortName := parts[len(parts)-1]

			// Config: projects/{project}/instanceConfigs/{config}
			configParts := strings.Split(i.Config, "/")
			shortConfig := configParts[len(configParts)-1]

			instances = append(instances, Instance{
				Name:            shortName,
				DisplayName:     i.DisplayName,
				ProjectID:       projectID,
				Config:          shortConfig,
				State:           i.State,
				NodeCount:       int(i.NodeCount),
				ProcessingUnits: int(i.ProcessingUnits),
				Labels:          i.Labels,
			})
		}
		return nil
	})
	return instances, err
}
