package gce

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/yogirk/tgcp/internal/ui/components"
)

func (s *Service) updateTable(instances []Instance) {
	rows := make([]table.Row, len(instances))
	for i, inst := range instances {
		rows[i] = InstanceToRow(inst)
	}
	s.table.SetRows(rows)
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
