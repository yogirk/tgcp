package gce

import (
	"time"

	"github.com/yogirk/tgcp/internal/demo"
)

// demoInstance is the JSON DTO for fixture data. Times are encoded as
// integer "days ago" values so the fixture stays evergreen without needing
// periodic date updates.
type demoInstance struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Zone           string   `json:"zone"`
	State          string   `json:"state"`
	MachineType    string   `json:"machineType"`
	InternalIP     string   `json:"internalIP"`
	ExternalIP     string   `json:"externalIP"`
	CreatedDaysAgo int      `json:"createdDaysAgo"`
	Tags           []string `json:"tags"`
	OSImage        string   `json:"osImage"`
	Disks          []Disk   `json:"disks"`
}

func loadDemoInstances() []Instance {
	var fixtures []demoInstance
	demo.MustLoad("gce", &fixtures)

	now := time.Now()
	out := make([]Instance, len(fixtures))
	for i, f := range fixtures {
		out[i] = Instance{
			ID:           f.ID,
			Name:         f.Name,
			Zone:         f.Zone,
			State:        InstanceState(f.State),
			MachineType:  f.MachineType,
			InternalIP:   f.InternalIP,
			ExternalIP:   f.ExternalIP,
			CreationTime: now.AddDate(0, 0, -f.CreatedDaysAgo),
			Tags:         f.Tags,
			Disks:        f.Disks,
			OSImage:      f.OSImage,
		}
	}
	return out
}
