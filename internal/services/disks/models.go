package disks

import "strings"

type Disk struct {
	Name                string
	Zone                string
	SizeGb              int64
	Type                string // e.g., pd-standard, pd-ssd, pd-balanced
	Status              string // READY, CREATING, RESTORING
	LastAttachTimestamp string
	Users               []string // Links to instances attached to this disk
	SourceImage         string   // Source image if boot disk
}

// IsOrphan returns true if the disk is not attached to any instance
func (d Disk) IsOrphan() bool {
	return len(d.Users) == 0
}

// ShortType returns a cleaner disk type string
func (d Disk) ShortType() string {
	// "projects/projectId/zones/zone/diskTypes/pd-standard" -> "pd-standard"
	parts := strings.Split(d.Type, "/")
	return parts[len(parts)-1]
}
