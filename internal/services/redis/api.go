package redis

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/api/redis/v1"
)

type Client struct {
	service *redis.Service
}

func NewClient(ctx context.Context) (*Client, error) {
	svc, err := redis.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("redis client: %w", err)
	}
	return &Client{service: svc}, nil
}

func (c *Client) ListInstances(projectID string) ([]Instance, error) {
	var instances []Instance
	// Aggregated list works best to find across regions
	parent := fmt.Sprintf("projects/%s/locations/-", projectID)

	err := c.service.Projects.Locations.Instances.List(parent).Pages(context.Background(), func(page *redis.ListInstancesResponse) error {
		for _, i := range page.Instances {
			// Name format: projects/{project}/locations/{location}/instances/{instance_id}
			parts := strings.Split(i.Name, "/")
			shortName := parts[len(parts)-1]
			location := ""
			if len(parts) > 3 {
				location = parts[len(parts)-3]
			}

			// Network: projects/{project}/global/networks/{network}
			netParts := strings.Split(i.AuthorizedNetwork, "/")
			network := i.AuthorizedNetwork
			if len(netParts) > 0 {
				network = netParts[len(netParts)-1]
			}

			instances = append(instances, Instance{
				Name:              shortName,
				DisplayName:       i.DisplayName,
				ProjectID:         projectID,
				Location:          location,
				Tier:              i.Tier,
				MemorySizeGb:      int(i.MemorySizeGb),
				RedisVersion:      i.RedisVersion,
				Host:              i.Host,
				Port:              int(i.Port),
				State:             i.State,
				AuthorizedNetwork: network,
			})
		}
		return nil
	})
	return instances, err
}
