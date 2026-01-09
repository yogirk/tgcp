package dataflow

import (
	"context"
	"fmt"

	dataflow "google.golang.org/api/dataflow/v1b3"
)

type Client struct {
	service *dataflow.Service
}

func NewClient(ctx context.Context) (*Client, error) {
	svc, err := dataflow.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("dataflow client: %w", err)
	}
	return &Client{service: svc}, nil
}

func (c *Client) ListJobs(projectID string) ([]Job, error) {
	var jobs []Job
	// Dataflow is regional, but has an aggregated list "jobs.aggregated" in v1b3?
	// Actually projects.jobs.aggregatedList exists.

	call := c.service.Projects.Jobs.Aggregated(projectID)
	err := call.Pages(context.Background(), func(page *dataflow.ListJobsResponse) error {
		for _, j := range page.Jobs {
			// Clean up state string "JOB_STATE_RUNNING" -> "RUNNING"
			// Clean up type "JOB_TYPE_STREAMING" -> "STREAMING"

			jobs = append(jobs, Job{
				ID:         j.Id,
				Name:       j.Name,
				Type:       j.Type,
				State:      j.CurrentState,
				CreateTime: j.CreateTime,
				Location:   j.Location,
			})
		}
		return nil
	})
	return jobs, err
}
