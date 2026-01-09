package components

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rk/tgcp/internal/styles"
)

type HomeMenuModel struct {
	Items     []ServiceItem
	Cursor    int
	IsFocused bool
}

func NewHomeMenu() HomeMenuModel {
	return HomeMenuModel{
		Items: []ServiceItem{
			{Name: "Overview", ShortName: "overview", Active: true},
			{Name: "Google Compute Engine (GCE)", ShortName: "gce"},
			{Name: "Cloud SQL", ShortName: "sql"},
			{Name: "Identity & Access Management (IAM)", ShortName: "iam"},
			{Name: "Cloud Run", ShortName: "run"},
			{Name: "Cloud Storage (GCS)", ShortName: "gcs"},
			{Name: "BigQuery", ShortName: "bq"},
			{Name: "Networking", ShortName: "net"},
		},
		Cursor:    0,
		IsFocused: true,
	}
}

func (m HomeMenuModel) Init() tea.Cmd {
	return nil
}

func (m HomeMenuModel) Update(msg tea.Msg) (HomeMenuModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "down", "j":
			if m.Cursor < len(m.Items)-1 {
				m.Cursor++
			}
		}
	}
	return m, nil
}

func (m HomeMenuModel) View() string {
	var items []string

	for i, item := range m.Items {
		name := item.Name
		if item.IsComing {
			name += " [Coming Soon]"
		}

		isSelected := m.Cursor == i

		var renderedItem string
		if isSelected {
			// Selected Item: Highlighted with a pointer
			renderedItem = styles.SelectedItemStyle.Copy().
				UnsetBorderLeft().
				Render("👉 " + name)
		} else {
			// Unselected Item: Dimmed
			style := styles.UnselectedItemStyle.Copy().
				PaddingLeft(2) // indent to match pointer

			if item.IsComing {
				style = style.Foreground(styles.ColorSubtext)
			}
			renderedItem = style.Render(name)
		}
		items = append(items, renderedItem)
	}

	// Join items with newlines
	listContent := lipgloss.JoinVertical(lipgloss.Left, items...)

	// Title
	title := styles.HeaderStyle.Render("📦 Available Services")

	// Combine Title + Content
	content := lipgloss.JoinVertical(lipgloss.Left, title, "", listContent)

	// Wrap in a card/box
	menuBox := styles.BoxStyle.Copy().
		BorderForeground(styles.ColorSubtext).
		Padding(1, 2).
		Render(content)

	return menuBox
}

func (m HomeMenuModel) SelectedItem() ServiceItem {
	return m.Items[m.Cursor]
}
