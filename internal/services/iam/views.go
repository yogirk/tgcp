package iam

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/yogirk/tgcp/internal/styles"
	"github.com/yogirk/tgcp/internal/ui/components"
)

func (s *Service) renderServiceAccountsList() string {
	breadcrumb := components.Breadcrumb(
		fmt.Sprintf("Project %s", s.projectID),
		s.Name(),
		"Service Accounts",
	)
	return lipgloss.JoinVertical(
		lipgloss.Left,
		breadcrumb,
		s.table.View(),
	)
}

func (s *Service) renderDetailView() string {
	if s.selectedAccount == nil {
		return "No account selected"
	}

	// Breadcrumb
	header := components.Breadcrumb(
		fmt.Sprintf("Project %s", s.projectID),
		s.Name(),
		"Service Accounts",
		s.selectedAccount.DisplayName,
	)

	content := lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		components.DetailCard(components.DetailCardOpts{
			Title: "Service Account Details",
			Rows: []components.KeyValue{
				{Key: "Display Name", Value: s.selectedAccount.DisplayName},
				{Key: "Email", Value: s.selectedAccount.Email},
				{Key: "Unique ID", Value: s.selectedAccount.UniqueID},
				{Key: "Status", Value: activeStatus(s.selectedAccount.Disabled)},
				{Key: "Description", Value: s.selectedAccount.Description},
			},
		}),
	)

	return content
}

func activeStatus(disabled bool) string {
	if disabled {
		return lipgloss.NewStyle().Foreground(styles.ColorError).Render("Disabled")
	}
	return lipgloss.NewStyle().Foreground(styles.ColorSuccess).Render("Active")
}
