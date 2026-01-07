package components

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type TableModel struct {
	Table table.Model
}

func NewTable(columns []table.Column, rows []table.Row) TableModel {
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("62")). // Purple selection
		Bold(false)
	t.SetStyles(s)

	return TableModel{Table: t}
}

func (m *TableModel) Update(msg tea.Msg) (TableModel, tea.Cmd) {
	var cmd tea.Cmd
	m.Table, cmd = m.Table.Update(msg)
	return *m, cmd
}

func (m TableModel) View() string {
	baseStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240"))

	return baseStyle.Render(m.Table.View())
}

func (m *TableModel) SetRows(rows []table.Row) {
	m.Table.SetRows(rows)
}

func (m *TableModel) SetHeight(h int) {
	m.Table.SetHeight(h)
}
