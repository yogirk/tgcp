package cloudrun

import (
	"context"
	"fmt"

	"google.golang.org/api/cloudfunctions/v2"
	"google.golang.org/api/option"
	run "google.golang.org/api/run/v1"
)

type Client struct {
	service   *run.APIService
	functions *cloudfunctions.Service
}

func NewClient(ctx context.Context) (*Client, error) {
	opts := []option.ClientOption{option.WithScopes(run.CloudPlatformScope)}

	svc, err := run.NewService(ctx, opts...)
	if err != nil {
		return nil, err
	}

	funcSvc, err := cloudfunctions.NewService(ctx, opts...)
	if err != nil {
		return nil, err
	}

	return &Client{service: svc, functions: funcSvc}, nil
}

func (c *Client) ListServices(projectID string) ([]RunService, error) {
	// List services across all locations ("-")
	parent := fmt.Sprintf("projects/%s/locations/-", projectID)

	resp, err := c.service.Projects.Locations.Services.List(parent).Do()
	if err != nil {
		return nil, err
	}

	var services []RunService
	for _, item := range resp.Items {
		// Parse useful information
		name := item.Metadata.Name
		// Name is often fully qualified, let's extract the simple name if needed?
		// But usually Metadata.Name IS the simple name in the k8s object,
		// but check if it returns standard k8s object.
		// Actually run/v1 returns a Service object where Metadata.Name is usually just "my-service".

		region := "global"
		if item.Metadata.Labels != nil {
			if loc, ok := item.Metadata.Labels["cloud.googleapis.com/location"]; ok {
				region = loc
			}
		}

		status := StatusUnknown
		url := ""
		if item.Status != nil {
			url = item.Status.Url
			for _, cond := range item.Status.Conditions {
				if cond.Type == "Ready" {
					if cond.Status == "True" {
						status = StatusReady
					} else if cond.Status == "False" {
						status = StatusFailed
					}
					break
				}
			}
		}

		services = append(services, RunService{
			Name:   name,
			Region: region,
			URL:    url,
			Status: status,
		})
	}
	return services, nil
}
