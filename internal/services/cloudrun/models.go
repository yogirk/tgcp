package cloudrun

import "time"

type ServiceStatus string

const (
	StatusReady   ServiceStatus = "Ready"
	StatusFailed  ServiceStatus = "Failed"
	StatusUnknown ServiceStatus = "Unknown"
)

type RunService struct {
	Name         string
	Region       string
	URL          string
	Status       ServiceStatus
	LastModified time.Time
}
