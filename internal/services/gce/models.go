package gce

import "time"

// InstanceState represents the status of a VM
type InstanceState string

const (
	StateRunning    InstanceState = "RUNNING"
	StateStopped    InstanceState = "STOPPED"
	StateTerminated InstanceState = "TERMINATED"
	StateOther      InstanceState = "OTHER"
)

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
}
