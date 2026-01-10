package cloudsql

import (
	"fmt"
	"strings"

	"github.com/yogirk/tgcp/internal/styles"
	"github.com/yogirk/tgcp/internal/ui/components"
)

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

	card := components.DetailCard(components.DetailCardOpts{
		Title: "Instance Details",
		Rows: []components.KeyValue{
			{Key: "Name", Value: i.Name},
			{Key: "State", Value: renderState(i.State)},
			{Key: "Database Version", Value: i.DatabaseVersion},
			{Key: "Region", Value: i.Region},
			{Key: "Tier", Value: i.Tier},
			{Key: "Storage (GB)", Value: fmt.Sprintf("%d", i.StorageGB)},
			{Key: "Auto Backup", Value: fmt.Sprintf("%v", i.AutoBackup)},
			{Key: "Activation Policy", Value: i.Activation},
			{Key: "Primary IP", Value: i.PrimaryIP},
			{Key: "Connection Name", Value: i.ConnectionName},
		},
		FooterHint: "s Start | x Stop | q Back",
	})

	doc.WriteString(card)

	return doc.String()
}

func (s *Service) renderConfirmation() string {
	if s.selectedInstance == nil {
		return "Error: No instance selected"
	}

	return components.RenderConfirmation(s.pendingAction, s.selectedInstance.Name, "instance")
}

func renderState(state InstanceState) string {
	str := string(state)
	switch state {
	case StateRunnable:
		return styles.SuccessStyle.Render("● " + str)
	case StateSuspended, StatePending:
		return styles.WarningStyle.Render("● " + str)
	case StateFailed:
		return styles.ErrorStyle.Render("● " + str)
	default:
		return str
	}
}
