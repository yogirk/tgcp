package iam

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/yogirk/tgcp/internal/styles"
)

func (s *Service) renderServiceAccountsList() string {
	return styles.BaseStyle.Render(s.table.View())
}

func (s *Service) renderDetailView() string {
	if s.selectedAccount == nil {
		return "No account selected"
	}

	// Breadcrumb
	header := styles.SubtleStyle.Render(fmt.Sprintf("IAM > Service Accounts > %s", s.selectedAccount.DisplayName))

	content := lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		fmtKeyVal("Display Name", s.selectedAccount.DisplayName),
		fmtKeyVal("Email", s.selectedAccount.Email),
		fmtKeyVal("Unique ID", s.selectedAccount.UniqueID),
		fmtKeyVal("Status", activeStatus(s.selectedAccount.Disabled)),
		fmtKeyVal("Description", s.selectedAccount.Description),
	)

	return styles.BoxStyle.Render(content)
}

func fmtKeyVal(key, val string) string {
	return lipgloss.NewStyle().Bold(true).Render(key+": ") + val
}

func activeStatus(disabled bool) string {
	if disabled {
		return lipgloss.NewStyle().Foreground(styles.ColorError).Render("Disabled")
	}
	return lipgloss.NewStyle().Foreground(styles.ColorSuccess).Render("Active")
}
