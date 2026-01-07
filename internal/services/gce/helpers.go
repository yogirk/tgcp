package gce

import (
	"strings"

	"github.com/charmbracelet/bubbles/table"
)

func (s *Service) updateTable(instances []Instance) {
	rows := make([]table.Row, len(instances))
	for i, inst := range instances {
		rows[i] = InstanceToRow(inst)
	}
	s.table.SetRows(rows)
	s.table.SetCursor(0)
}

// filterInstances updates the filtered list based on query
func (s *Service) filterInstances(query string) {
	if query == "" {
		s.filtered = nil
		s.updateTable(s.instances)
		return
	}

	var matches []Instance
	for _, inst := range s.instances {
		if contains(inst.Name, query) || contains(inst.Zone, query) {
			matches = append(matches, inst)
		}
	}
	s.filtered = matches
	s.updateTable(s.filtered)
}

// getCurrentInstances returns the currently visible instances
func (s *Service) getCurrentInstances() []Instance {
	if s.filtered != nil || s.filterInput.Value() != "" {
		return s.filtered
	}
	return s.instances
}

func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
