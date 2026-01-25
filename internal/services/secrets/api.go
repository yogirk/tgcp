package secrets

import (
	"context"
	"fmt"
	"strings"
	"time"

	secretmanager "google.golang.org/api/secretmanager/v1"
)

// Client wraps the Secret Manager API
type Client struct {
	service *secretmanager.Service
}

// NewClient creates a new Secret Manager client
func NewClient(ctx context.Context) (*Client, error) {
	svc, err := secretmanager.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create secretmanager service: %w", err)
	}
	return &Client{service: svc}, nil
}

// ListSecrets returns all secrets in the project
func (c *Client) ListSecrets(projectID string) ([]Secret, error) {
	parent := fmt.Sprintf("projects/%s", projectID)
	var secrets []Secret

	req := c.service.Projects.Secrets.List(parent)
	err := req.Pages(context.Background(), func(resp *secretmanager.ListSecretsResponse) error {
		for _, s := range resp.Secrets {
			secret := Secret{
				FullName:    s.Name,
				Name:        extractSecretName(s.Name),
				Labels:      s.Labels,
				Replication: formatReplication(s.Replication),
			}

			// Parse create time
			if s.CreateTime != "" {
				if t, err := time.Parse(time.RFC3339Nano, s.CreateTime); err == nil {
					secret.CreateTime = t
				}
			}

			secrets = append(secrets, secret)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}

	return secrets, nil
}

// GetSecret returns details for a specific secret including version count
func (c *Client) GetSecret(secretName string) (*Secret, error) {
	resp, err := c.service.Projects.Secrets.Get(secretName).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}

	secret := &Secret{
		FullName:    resp.Name,
		Name:        extractSecretName(resp.Name),
		Labels:      resp.Labels,
		Replication: formatReplication(resp.Replication),
	}

	if resp.CreateTime != "" {
		if t, err := time.Parse(time.RFC3339Nano, resp.CreateTime); err == nil {
			secret.CreateTime = t
		}
	}

	return secret, nil
}

// ListVersions returns all versions for a secret
func (c *Client) ListVersions(secretName string) ([]SecretVersion, error) {
	var versions []SecretVersion

	req := c.service.Projects.Secrets.Versions.List(secretName)
	err := req.Pages(context.Background(), func(resp *secretmanager.ListSecretVersionsResponse) error {
		for _, v := range resp.Versions {
			version := SecretVersion{
				FullName: v.Name,
				Name:     extractVersionNumber(v.Name),
				State:    v.State,
			}

			if v.CreateTime != "" {
				if t, err := time.Parse(time.RFC3339Nano, v.CreateTime); err == nil {
					version.CreateTime = t
				}
			}

			versions = append(versions, version)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list versions: %w", err)
	}

	return versions, nil
}

// extractSecretName extracts the secret name from the full resource name
// e.g., "projects/my-project/secrets/my-secret" -> "my-secret"
func extractSecretName(fullName string) string {
	parts := strings.Split(fullName, "/")
	if len(parts) >= 4 {
		return parts[3]
	}
	return fullName
}

// extractVersionNumber extracts version number from full resource name
// e.g., "projects/my-project/secrets/my-secret/versions/1" -> "1"
func extractVersionNumber(fullName string) string {
	parts := strings.Split(fullName, "/")
	if len(parts) >= 6 {
		return parts[5]
	}
	return fullName
}

// formatReplication formats the replication config for display
func formatReplication(r *secretmanager.Replication) string {
	if r == nil {
		return "unknown"
	}
	if r.Automatic != nil {
		return "automatic"
	}
	if r.UserManaged != nil && len(r.UserManaged.Replicas) > 0 {
		var regions []string
		for _, replica := range r.UserManaged.Replicas {
			regions = append(regions, replica.Location)
		}
		return strings.Join(regions, ", ")
	}
	return "unknown"
}
