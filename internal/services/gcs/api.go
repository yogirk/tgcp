package gcs

import (
	"context"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

type Client struct {
	client *storage.Client
}

func NewClient(ctx context.Context) (*Client, error) {
	c, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	return &Client{client: c}, nil
}

func (c *Client) ListBuckets(projectID string) ([]Bucket, error) {
	var buckets []Bucket
	it := c.client.Buckets(context.Background(), projectID)
	for {
		battrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		buckets = append(buckets, Bucket{
			Name:         battrs.Name,
			Location:     battrs.Location,
			StorageClass: battrs.StorageClass,
			Created:      battrs.Created,
		})
	}
	return buckets, nil
}

func (c *Client) ListObjects(bucket, prefix string) ([]Object, error) {
	var objects []Object
	it := c.client.Bucket(bucket).Objects(context.Background(), &storage.Query{
		Prefix:    prefix,
		Delimiter: "/",
	})

	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		if attrs.Prefix != "" {
			// It's a folder
			objects = append(objects, Object{
				Name: attrs.Prefix,
				Type: "Folder",
			})
		} else {
			// It's a file
			objects = append(objects, Object{
				Name:    attrs.Name,
				Size:    attrs.Size,
				Updated: attrs.Updated,
				Type:    attrs.ContentType,
			})
		}
	}
	return objects, nil
}
