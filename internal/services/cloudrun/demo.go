package cloudrun

import (
	"time"

	"github.com/yogirk/tgcp/internal/demo"
)

type demoRunService struct {
	Name                string `json:"name"`
	Region              string `json:"region"`
	URL                 string `json:"url"`
	Status              string `json:"status"`
	LastModifiedDaysAgo int    `json:"lastModifiedDaysAgo"`
}

func loadDemoRunServices() []RunService {
	var fixtures []demoRunService
	demo.MustLoad("cloudrun", &fixtures)

	now := time.Now()
	out := make([]RunService, len(fixtures))
	for i, f := range fixtures {
		out[i] = RunService{
			Name:         f.Name,
			Region:       f.Region,
			URL:          f.URL,
			Status:       ServiceStatus(f.Status),
			LastModified: now.AddDate(0, 0, -f.LastModifiedDaysAgo),
		}
	}
	return out
}
