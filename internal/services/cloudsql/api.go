package cloudsql

import (
	"context"
	"fmt"

	"github.com/rk/tgcp/internal/core"
	"google.golang.org/api/option"
	sqladmin "google.golang.org/api/sqladmin/v1beta4"
)

// Client wraps the Cloud SQL Admin API
type Client struct {
	service *sqladmin.Service
}

// NewClient creates a new Cloud SQL client
func NewClient(ctx context.Context) (*Client, error) {
	httpClient, err := core.NewHTTPClient(ctx, sqladmin.SqlserviceAdminScope)
	if err != nil {
		return nil, fmt.Errorf("failed to create http client: %w", err)
	}

	svc, err := sqladmin.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("failed to create sql client: %w", err)
	}

	return &Client{service: svc}, nil
}

// ListInstances fetches all Cloud SQL instances in the project
func (c *Client) ListInstances(projectID string) ([]Instance, error) {
	resp, err := c.service.Instances.List(projectID).Do()
	if err != nil {
		return nil, fmt.Errorf("cloud sql api error: %w", err)
	}

	var instances []Instance
	for _, item := range resp.Items {
		// Find Primary IP
		primaryIP := "N/A"
		for _, ip := range item.IpAddresses {
			if ip.Type == "PRIMARY" {
				primaryIP = ip.IpAddress
				break
			}
		}

		inst := Instance{
			Name:            item.Name,
			ProjectID:       item.Project,
			Region:          item.Region,
			DatabaseVersion: item.DatabaseVersion,
			State:           InstanceState(item.State),
			PrimaryIP:       primaryIP,
			ConnectionName:  item.ConnectionName,
		}

		// Detailed mapping
		if item.Settings != nil {
			inst.Tier = item.Settings.Tier
			inst.Activation = item.Settings.ActivationPolicy
			if item.Settings.DataDiskSizeGb > 0 {
				inst.StorageGB = item.Settings.DataDiskSizeGb
			}
			if item.Settings.BackupConfiguration != nil {
				inst.AutoBackup = item.Settings.BackupConfiguration.Enabled
			}
		}

		instances = append(instances, inst)
	}

	return instances, nil
}

// StopInstance stops a Cloud SQL instance by setting activation policy to NEVER
func (c *Client) StopInstance(projectID, name string) error {
	rb := &sqladmin.DatabaseInstance{
		Settings: &sqladmin.Settings{
			ActivationPolicy: "NEVER",
		},
	}
	_, err := c.service.Instances.Patch(projectID, name, rb).Do()
	return err
}

// StartInstance starts a Cloud SQL instance by setting activation policy to ALWAYS
func (c *Client) StartInstance(projectID, name string) error {
	rb := &sqladmin.DatabaseInstance{
		Settings: &sqladmin.Settings{
			ActivationPolicy: "ALWAYS",
		},
	}
	_, err := c.service.Instances.Patch(projectID, name, rb).Do()
	return err
}
