package gcs

import "time"

type Bucket struct {
	Name         string
	Location     string
	StorageClass string
	Created      time.Time
}

type Object struct {
	Name    string
	Size    int64
	Updated time.Time
	Type    string // "Folder" or ContentType
}
