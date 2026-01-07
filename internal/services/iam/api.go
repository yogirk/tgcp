package iam

import (
	"context"
	"fmt"

	"github.com/rk/tgcp/internal/core"
	"google.golang.org/api/iam/v1"
	"google.golang.org/api/option"
)

// Client handles IAM API interactions
type Client struct {
	service *iam.Service
}

// NewClient creates a new IAM API client
func NewClient(ctx context.Context) (*Client, error) {
	httpClient, err := core.NewHTTPClient(ctx, iam.CloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("failed to create http client: %w", err)
	}

	service, err := iam.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("failed to create iam service: %w", err)
	}
	return &Client{service: service}, nil
}

// ListServiceAccounts lists all service accounts in the project
// resource should be "projects/PROJECT_ID"
func (c *Client) ListServiceAccounts(projectID string) ([]ServiceAccount, error) {
	resource := "projects/" + projectID
	resp, err := c.service.Projects.ServiceAccounts.List(resource).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list service accounts: %w", err)
	}

	var accounts []ServiceAccount
	for _, acc := range resp.Accounts {
		accounts = append(accounts, ServiceAccount{
			Name:        acc.Name,
			Email:       acc.Email,
			DisplayName: acc.DisplayName,
			Description: acc.Description,
			Disabled:    acc.Disabled,
			UniqueID:    acc.UniqueId,
		})
	}
	return accounts, nil
}
