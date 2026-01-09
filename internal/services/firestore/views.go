package firestore

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/rk/tgcp/internal/styles"
)

func (s *Service) View() string {
	if s.loading && len(s.dbs) == 0 {
		return "Loading Firestore Databases..."
	}
	if s.err != nil {
		return fmt.Sprintf("Error: %v", s.err)
	}

	if s.viewState == ViewDetail {
		return s.renderDetailView()
	}

	return s.table.View()
}

func (s *Service) renderDetailView() string {
	db := s.selectedDB
	if db == nil {
		return ""
	}

	header := lipgloss.JoinHorizontal(lipgloss.Left,
		styles.BaseStyle.Foreground(styles.ColorPrimary).Render("🔥 "),
		styles.HeaderStyle.Render(fmt.Sprintf("DB: %s", db.Name)),
	)

	details := fmt.Sprintf(
		"Type: %s\nLocation: %s\nCreated: %s\nUID: %s",
		strings.Replace(db.Type, "FIRESTORE_", "", 1),
		db.Location,
		db.CreateTime,
		db.Uid,
	)

	box := styles.BoxStyle.Copy().Width(80).Render(
		lipgloss.JoinVertical(lipgloss.Left, header, " ", details),
	)
	return box
}
