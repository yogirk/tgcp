package gce

import (
	"testing"
)

func TestEstimateCost(t *testing.T) {
	tests := []struct {
		name        string
		machineType string
		zone        string
		disks       []Disk
		want        string
	}{
		{
			"VM only",
			"e2-medium", // 0.033
			"us-central1-a",
			nil,
			"$0.033/hr",
		},
		{
			"VM with Standard Disk",
			"e2-medium", // 0.033
			"us-central1-a",
			[]Disk{{SizeGB: 100, Type: "pd-standard"}}, // 100 * 0.04 / 730 = 0.005479... -> Total ~0.038
			"$0.038/hr",
		},
		{
			"VM with SSD Disk",
			"e2-medium", // 0.033
			"us-central1-a",
			[]Disk{{SizeGB: 100, Type: "pd-ssd"}}, // 100 * 0.17 / 730 = 0.02328... -> Total 0.033+0.023=0.056
			"$0.056/hr",
		},
		{
			"Unknown VM",
			"unknown-vm",
			"us-central1-a",
			nil,
			"N/A",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EstimateCost(tt.machineType, tt.zone, tt.disks); got != tt.want {
				t.Errorf("EstimateCost() = %v, want %v", got, tt.want)
			}
		})
	}
}
