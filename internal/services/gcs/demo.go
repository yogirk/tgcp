package gcs

import (
	"time"

	"github.com/yogirk/tgcp/internal/demo"
)

type demoBucket struct {
	Name           string `json:"name"`
	Location       string `json:"location"`
	StorageClass   string `json:"storageClass"`
	CreatedDaysAgo int    `json:"createdDaysAgo"`
}

type demoObject struct {
	Name           string `json:"name"`
	Size           int64  `json:"size"`
	UpdatedDaysAgo int    `json:"updatedDaysAgo"`
	Type           string `json:"type"`
}

func loadDemoBuckets() []Bucket {
	var fixtures []demoBucket
	demo.MustLoad("gcs_buckets", &fixtures)

	now := time.Now()
	out := make([]Bucket, len(fixtures))
	for i, f := range fixtures {
		out[i] = Bucket{
			Name:         f.Name,
			Location:     f.Location,
			StorageClass: f.StorageClass,
			Created:      now.AddDate(0, 0, -f.CreatedDaysAgo),
		}
	}
	return out
}

func loadDemoObjects() []Object {
	var fixtures []demoObject
	demo.MustLoad("gcs_objects", &fixtures)

	now := time.Now()
	out := make([]Object, len(fixtures))
	for i, f := range fixtures {
		out[i] = Object{
			Name:    f.Name,
			Size:    f.Size,
			Updated: now.AddDate(0, 0, -f.UpdatedDaysAgo),
			Type:    f.Type,
		}
	}
	return out
}
