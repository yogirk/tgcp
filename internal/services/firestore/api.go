package firestore

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/api/firestore/v1"
)

type Client struct {
	service *firestore.Service
}

func NewClient(ctx context.Context) (*Client, error) {
	svc, err := firestore.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("firestore client: %w", err)
	}
	return &Client{service: svc}, nil
}

func (c *Client) ListDatabases(projectID string) ([]Database, error) {
	var dbs []Database
	parent := fmt.Sprintf("projects/%s", projectID)

	// firestore/v1 projects.databases.list
	call := c.service.Projects.Databases.List(parent)
	resp, err := call.Do() // Basic Do() as iterating pages manually is verbose, and DB count is low (usually 1)
	if err != nil {
		return nil, err
	}

	for _, db := range resp.Databases {
		// Name: projects/{project}/databases/{database_id}
		parts := strings.Split(db.Name, "/")
		shortName := parts[len(parts)-1]

		dbs = append(dbs, Database{
			Name:      shortName,
			ProjectID: projectID,
			Location:  db.LocationId,
			Type:      db.Type,
			State:     "READY", // API v1 Database object doesn't always show state clearly in struct? Checking docs...
			// Actually Database object has `Uid`, `CreateTime`, `UpdateTime`, `LocationId`, `Type`, `ConcurrencyMode`, etc.
			// "State" key might be missing in basic v1 struct or it's implicitly Active.
			CreateTime: db.CreateTime,
			Uid:        db.Uid,
		})
	}
	return dbs, nil
}
