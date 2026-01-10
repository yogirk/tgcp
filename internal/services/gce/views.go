package gce

import (
	"fmt"
	"strings"
	"time"

	"github.com/yogirk/tgcp/internal/styles"
	"github.com/yogirk/tgcp/internal/ui/components"
)

// renderDetailView renders the details of a single instance
func (s *Service) renderDetailView() string {
	if s.selectedInstance == nil {
		return "No instance selected"
	}
	i := s.selectedInstance

	doc := strings.Builder{}

	// Breadcrumb
	doc.WriteString(components.Breadcrumb(
		fmt.Sprintf("Project %s", s.projectID),
		s.Name(),
		"Instances",
		i.Name,
	))
	doc.WriteString("\n\n")

	// Instance Details Section
	// Calculate Total Disk Size
	var totalDisk int64
	for _, d := range i.Disks {
		totalDisk += d.SizeGB
	}

	// Calculate Age
	// Simple duration format: Xd Yh
	age := time.Since(i.CreationTime)
	days := int(age.Hours() / 24)
	ageStr := fmt.Sprintf("%d days ago", days)
	if days == 0 {
		hours := int(age.Hours())
		ageStr = fmt.Sprintf("%d hours ago", hours)
	}

	card := components.DetailCard(components.DetailCardOpts{
		Title: "Instance Details",
		Rows: []components.KeyValue{
			{Key: "Name", Value: i.Name},
			{Key: "Status", Value: renderStatus(i.State)},
			{Key: "Zone", Value: i.Zone},
			{Key: "Machine Type", Value: i.MachineType},
			{Key: "OS Image", Value: i.OSImage},
			{Key: "Disk Size", Value: fmt.Sprintf("%d GB", totalDisk)},
			{Key: "Created", Value: ageStr},
			{Key: "Estimated Cost", Value: EstimateCost(i.MachineType, i.Zone, i.Disks)},
			{Key: "Internal IP", Value: i.InternalIP},
			{Key: "External IP", Value: i.ExternalIP},
		},
		FooterHint: "s Start | x Stop | h SSH | q Back",
	})

	doc.WriteString(card)

	return doc.String()
}

func renderStatus(state InstanceState) string {
	return components.RenderStatus(string(state))
}

// renderConfirmation renders a confirmation dialog
func (s *Service) renderConfirmation() string {
	if s.selectedInstance == nil {
		return "Error: No instance selected"
	}

	return components.RenderConfirmation(s.pendingAction, s.selectedInstance.Name, "instance")
}

// renderListView renders the main instance table
func (s *Service) renderListView() string {
	doc := strings.Builder{}

	// Breadcrumb + Filter Bar
	doc.WriteString(components.Breadcrumb(
		fmt.Sprintf("Project %s", s.projectID),
		s.Name(),
		"Instances",
	))
	doc.WriteString("\n")
	doc.WriteString(s.filter.View())
	doc.WriteString("\n")

	doc.WriteString(styles.BaseStyle.Render(s.table.View()))
	return doc.String()
}
