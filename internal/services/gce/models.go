package gce

import "time"

// InstanceState represents the status of a VM
type InstanceState string

const (
	StateRunning      InstanceState = "RUNNING"
	StateStopped      InstanceState = "STOPPED"
	StateTerminated   InstanceState = "TERMINATED"
	StateProvisioning InstanceState = "PROVISIONING"
	StateStaging      InstanceState = "STAGING"
	StateStopping     InstanceState = "STOPPING"
	StateSuspending   InstanceState = "SUSPENDING"
	StateRepairing    InstanceState = "REPAIRING"
	StateOther        InstanceState = "OTHER"
)

// Disk represents a GCE disk
type Disk struct {
	Name   string
	SizeGB int64
	Type   string // e.g. "pd-standard", "pd-ssd"
}

// Instance represents a simplified GCE VM
type Instance struct {
	ID           string
	Name         string
	Zone         string
	State        InstanceState
	MachineType  string
	InternalIP   string
	ExternalIP   string
	CreationTime time.Time
	Tags         []string
	Disks        []Disk
	OSImage      string
}
