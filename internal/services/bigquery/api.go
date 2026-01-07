package bigquery

import (
	"context"
	"fmt"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
)

type Client struct {
	client *bigquery.Client
}

func NewClient(ctx context.Context, projectID string) (*Client, error) {
	c, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create bigquery client: %w", err)
	}
	return &Client{client: c}, nil
}

func (c *Client) ListDatasets(projectID string) ([]Dataset, error) {
	var datasets []Dataset
	it := c.client.Datasets(context.Background())
	for {
		ds, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		// Fetch metadata for Location
		md, err := ds.Metadata(context.Background())
		if err != nil {
			// If error fetching metadata, we can still list ID?
			// Or just skip/log? Let's just list ID and unknown location
			datasets = append(datasets, Dataset{
				ID:        ds.DatasetID,
				ProjectID: ds.ProjectID,
				Location:  "UNKNOWN",
			})
			continue
		}

		datasets = append(datasets, Dataset{
			ID:        ds.DatasetID,
			ProjectID: ds.ProjectID,
			Location:  md.Location,
		})
	}
	return datasets, nil
}

func (c *Client) ListTables(datasetID string) ([]Table, error) {
	var tables []Table
	ds := c.client.Dataset(datasetID)
	it := ds.Tables(context.Background())
	for {
		t, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		// We need to fetch metadata to get NumRows etc.
		// NOTE: Fetching metadata for ALL tables might be slow properly.
		// For MVP, we might skip detailed metadata in the list or do parallel fetch?
		// Let's just return basic info for list and maybe fetch details strictly if needed.
		// Actually, `t` is *Table. The iterator returns basic info.
		// Let's assume for now we just want IDs.
		// Wait, user wants drill down.
		// Let's fetch metadata for the *Table* to get type.

		md, err := t.Metadata(context.Background())
		if err != nil {
			// Skip or handle? Let's just log or ignore?
			// Better to show it with unknown data.
			tables = append(tables, Table{
				ID:        t.TableID,
				DatasetID: datasetID,
				Type:      "UNKNOWN",
			})
			continue
		}

		tables = append(tables, Table{
			ID:         t.TableID,
			DatasetID:  datasetID,
			Type:       string(md.Type),
			NumRows:    md.NumRows,
			TotalBytes: md.NumBytes,
			LastMod:    md.LastModifiedTime,
		})
	}
	return tables, nil
}

func (c *Client) GetTableSchema(datasetID, tableID string) ([]SchemaField, error) {
	ds := c.client.Dataset(datasetID)
	t := ds.Table(tableID)
	md, err := t.Metadata(context.Background())
	if err != nil {
		return nil, err
	}

	var fields []SchemaField
	for _, f := range md.Schema {
		fields = append(fields, SchemaField{
			Name:        f.Name,
			Type:        string(f.Type),
			Mode:        "", // Will be populated shortly
			Description: f.Description,
		})
	}
	// Re-map mode correctly
	for i, f := range md.Schema {
		mode := "NULLABLE"
		if f.Required {
			mode = "REQUIRED"
		} else if f.Repeated {
			mode = "REPEATED"
		}
		fields[i].Mode = mode
	}

	return fields, nil
}
