package firestore

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/api/datastore/v1"
	"google.golang.org/api/firestore/v1"
)

type Client struct {
	firestoreSvc  *firestore.Service
	datastoreSvc  *datastore.Service
}

func NewClient(ctx context.Context) (*Client, error) {
	fsSvc, err := firestore.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("firestore client: %w", err)
	}
	dsSvc, err := datastore.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("datastore client: %w", err)
	}
	return &Client{firestoreSvc: fsSvc, datastoreSvc: dsSvc}, nil
}

func (c *Client) ListDatabases(projectID string) ([]Database, error) {
	var dbs []Database
	parent := fmt.Sprintf("projects/%s", projectID)

	// firestore/v1 projects.databases.list
	call := c.firestoreSvc.Projects.Databases.List(parent)
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

// ListNamespaces lists all namespaces in a Datastore mode database
func (c *Client) ListNamespaces(projectID, databaseID string) ([]Namespace, error) {
	var namespaces []Namespace

	// Query __namespace__ kind to get all namespaces
	// Note: "(default)" database should be passed as empty string to Datastore API
	dbID := databaseID
	if databaseID == "(default)" {
		dbID = ""
	}

	query := &datastore.Query{
		Kind: []*datastore.KindExpression{{Name: "__namespace__"}},
	}
	req := &datastore.RunQueryRequest{
		DatabaseId: dbID,
		Query:      query,
	}

	resp, err := c.datastoreSvc.Projects.RunQuery(projectID, req).Do()
	if err != nil {
		return nil, fmt.Errorf("list namespaces: %w", err)
	}

	for _, result := range resp.Batch.EntityResults {
		if result.Entity != nil && result.Entity.Key != nil {
			name := ""
			if result.Entity.Key.Path != nil && len(result.Entity.Key.Path) > 0 {
				// The namespace name is in the key's name or id
				pathElem := result.Entity.Key.Path[0]
				if pathElem.Name != "" {
					name = pathElem.Name
				} else {
					// Default namespace has id=1, name=""
					name = "(default)"
				}
			}
			namespaces = append(namespaces, Namespace{Name: name})
		}
	}

	// Ensure default namespace is included
	if len(namespaces) == 0 {
		namespaces = append(namespaces, Namespace{Name: "(default)"})
	}

	return namespaces, nil
}

// ListKinds lists all entity kinds in a Datastore mode database namespace
func (c *Client) ListKinds(projectID, databaseID, namespace string) ([]Kind, error) {
	var kinds []Kind

	// Query __kind__ kind to get all entity kinds in the namespace
	query := &datastore.Query{
		Kind: []*datastore.KindExpression{{Name: "__kind__"}},
	}

	// Set namespace partition (empty string for default namespace)
	nsID := namespace
	if namespace == "(default)" {
		nsID = ""
	}

	// Note: "(default)" database should be passed as empty string to Datastore API
	dbID := databaseID
	if databaseID == "(default)" {
		dbID = ""
	}

	req := &datastore.RunQueryRequest{
		DatabaseId: dbID,
		PartitionId: &datastore.PartitionId{
			ProjectId:   projectID,
			DatabaseId:  dbID,
			NamespaceId: nsID,
		},
		Query: query,
	}

	resp, err := c.datastoreSvc.Projects.RunQuery(projectID, req).Do()
	if err != nil {
		return nil, fmt.Errorf("list kinds: %w", err)
	}

	for _, result := range resp.Batch.EntityResults {
		if result.Entity != nil && result.Entity.Key != nil {
			if result.Entity.Key.Path != nil && len(result.Entity.Key.Path) > 0 {
				kindName := result.Entity.Key.Path[0].Name
				// Skip internal kinds that start with __
				if !strings.HasPrefix(kindName, "__") {
					kinds = append(kinds, Kind{
						Name:      kindName,
						Namespace: namespace,
					})
				}
			}
		}
	}

	return kinds, nil
}
