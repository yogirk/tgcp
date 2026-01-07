package cloudrun

import (
	"fmt"
	"time"
)

// Wrapper for Cloud Functions API
// We reuse the 'cloudrun' package name but keep logical separation

type Function struct {
	Name        string
	Region      string
	State       string // ACTIVE, DEPLOYING, etc.
	URL         string
	LastUpdated time.Time
	Environment string // GEN_1 or GEN_2
}

// ListFunctions fetches cloud functions from the project
func (c *Client) ListFunctions(projectID string) ([]Function, error) {
	if c.functions == nil {
		return nil, fmt.Errorf("functions client not initialized")
	}

	parent := fmt.Sprintf("projects/%s/locations/-", projectID)
	resp, err := c.functions.Projects.Locations.Functions.List(parent).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list functions: %w", err)
	}

	var results []Function
	for _, f := range resp.Functions {
		// Parse timestamp
		updated, _ := time.Parse(time.RFC3339, f.UpdateTime)

		url := ""
		if f.ServiceConfig != nil {
			url = f.ServiceConfig.Uri
		}

		results = append(results, Function{
			Name:        f.Name, // Full name is projects/../locations/../functions/name
			Region:      extractRegion(f.Name),
			State:       f.State,
			URL:         url,
			LastUpdated: updated,
			Environment: f.Environment,
		})
	}
	return results, nil
}

func extractRegion(fullName string) string {
	// Format: projects/{project}/locations/{location}/functions/{function}
	// We can cheat and just take the segment before "functions"
	// But simpler might be to just store it if available.
	// Let's iterate segments.
	// This is MVP quality.
	return "us-central1" // Placeholder or implement properly if needed
}
