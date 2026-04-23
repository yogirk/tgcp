package gce

import (
	"testing"
)

func TestInstanceToRow(t *testing.T) {
	tests := []struct {
		name     string
		instance Instance
		expected []string
	}{
		{
			name: "Running Instance",
			instance: Instance{
				ID:         "123",
				Name:       "test-vm",
				Zone:       "us-central1-a",
				State:      StateRunning,
				InternalIP: "10.0.0.2",
				ExternalIP: "34.1.2.3",
			},
			expected: []string{"test-vm", "🟢 RUNNING", "us-central1-a", "10.0.0.2", "34.1.2.3"},
		},
		{
			name: "Stopped Instance",
			instance: Instance{
				ID:         "456",
				Name:       "stopped-vm",
				Zone:       "us-east1-b",
				State:      StateStopped,
				InternalIP: "10.0.0.3",
				ExternalIP: "<none>",
			},
			expected: []string{"stopped-vm", "🔴 STOPPED", "us-east1-b", "10.0.0.3", "<none>"},
		},
		{
			name: "Terminated Instance",
			instance: Instance{
				ID:         "789",
				Name:       "terminated-vm",
				Zone:       "us-west1-c",
				State:      StateTerminated,
				InternalIP: "10.0.0.4",
				ExternalIP: "<none>",
			},
			expected: []string{"terminated-vm", "🔴 STOP", "us-west1-c", "10.0.0.4", "<none>"},
		},
		{
			name: "Provisioning Instance",
			instance: Instance{
				ID:         "101",
				Name:       "new-vm",
				Zone:       "us-central1-a",
				State:      StateProvisioning,
				InternalIP: "",
				ExternalIP: "",
			},
			expected: []string{"new-vm", "🔄 PROVISIONING", "us-central1-a", "", ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := InstanceToRow(tt.instance)
			
			if len(row) != len(tt.expected) {
				t.Fatalf("Expected row length %d, got %d", len(tt.expected), len(row))
			}

			for i, val := range row {
				if val != tt.expected[i] {
					t.Errorf("Column %d: expected %q, got %q", i, tt.expected[i], val)
				}
			}
		})
	}
}

func TestGetGCEColumns(t *testing.T) {
	cols := GetGCEColumns()
	
	if len(cols) != 5 {
		t.Errorf("Expected 5 columns, got %d", len(cols))
	}
	
	if cols[0].Title != "VM Name" {
		t.Errorf("Expected first column to be 'VM Name', got %q", cols[0].Title)
	}
}
