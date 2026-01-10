package gce

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/yogirk/tgcp/internal/ui/components"
)

func (s *Service) updateTable(instances []Instance) {
	rows := make([]table.Row, len(instances))
	for i, inst := range instances {
		status := string(inst.State)
		if status == "RUNNING" {
			status = "RUNNING"
		} else if status == "STOPPED" || status == "TERMINATED" {
			status = "STOPPED" // Simplify or keep original
		} else {
			status = string(inst.State)
		}

		rows[i] = table.Row{
			inst.Name,
			status,
			inst.Zone,
			inst.InternalIP,
			inst.ExternalIP,
			inst.ID,
		}
	}
	s.table.SetRows(rows)
	// s.table.SetCursor(0) // Don't reset cursor on every update, preserves selection on refresh
}

// getFilteredInstances returns filtered instances based on the query string
func (s *Service) getFilteredInstances(instances []Instance, query string) []Instance {
	if query == "" {
		return instances
	}
	return components.FilterSlice(instances, query, func(inst Instance, q string) bool {
		return components.ContainsMatch(inst.Name, inst.Zone, inst.InternalIP, inst.ExternalIP)(q)
	})
}
