package firestore

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/yogirk/tgcp/internal/ui/components"
)

func (s *Service) View() string {
	if s.err != nil {
		return components.RenderError(s.err, "Firestore", "Databases")
	}

	// Show spinner while loading
	if s.spinner.IsActive() {
		return s.spinner.View()
	}

	switch s.viewState {
	case ViewDetail:
		return s.renderDetailView()
	case ViewNamespaces:
		return s.renderNamespacesView()
	case ViewKinds:
		return s.renderKindsView()
	}

	// Default: List view
	var content strings.Builder
	content.WriteString(components.Breadcrumb(
		fmt.Sprintf("Project %s", s.projectID),
		s.Name(),
		"Databases",
	))
	content.WriteString("\n")
	content.WriteString(s.filter.View())
	content.WriteString("\n")
	content.WriteString(s.table.View())
	return content.String()
}

func (s *Service) renderDetailView() string {
	db := s.selectedDB
	if db == nil {
		return ""
	}

	breadcrumb := components.Breadcrumb(
		fmt.Sprintf("Project %s", s.projectID),
		s.Name(),
		"Databases",
		db.Name,
	)

	card := components.DetailCard(components.DetailCardOpts{
		Title: "Database Details",
		Rows: []components.KeyValue{
			{Key: "Name", Value: db.Name},
			{Key: "Type", Value: strings.Replace(db.Type, "FIRESTORE_", "", 1)},
			{Key: "Location", Value: db.Location},
			{Key: "Created", Value: db.CreateTime},
			{Key: "UID", Value: db.Uid},
		},
	})
	return lipgloss.JoinVertical(lipgloss.Left, breadcrumb, "", card)
}

func (s *Service) renderNamespacesView() string {
	if s.selectedDB == nil {
		return ""
	}

	var content strings.Builder
	content.WriteString(components.Breadcrumb(
		fmt.Sprintf("Project %s", s.projectID),
		s.Name(),
		s.selectedDB.Name+" (Datastore)",
		"Namespaces",
	))
	content.WriteString("\n\n")
	content.WriteString(s.nsTable.View())
	return content.String()
}

func (s *Service) renderKindsView() string {
	if s.selectedDB == nil || s.selectedNamespace == nil {
		return ""
	}

	var content strings.Builder
	content.WriteString(components.Breadcrumb(
		fmt.Sprintf("Project %s", s.projectID),
		s.Name(),
		s.selectedDB.Name+" (Datastore)",
		s.selectedNamespace.Name,
		"Kinds",
	))
	content.WriteString("\n\n")
	content.WriteString(s.kindTable.View())
	return content.String()
}
