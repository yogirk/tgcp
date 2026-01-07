package components

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type FilterModel struct {
	TextInput textinput.Model
	Active    bool
}

func NewFilter() FilterModel {
	ti := textinput.New()
	ti.Placeholder = "Filter..."
	ti.CharLimit = 50
	ti.Width = 30
	ti.Prompt = "/ "

	return FilterModel{
		TextInput: ti,
		Active:    false,
	}
}

func (m FilterModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m FilterModel) Update(msg tea.Msg) (FilterModel, tea.Cmd) {
	var cmd tea.Cmd
	m.TextInput, cmd = m.TextInput.Update(msg)
	return m, cmd
}

func (m FilterModel) View() string {
	if !m.Active {
		return ""
	}

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("214")). // Orange
		Padding(0, 1)

	return style.Render(m.TextInput.View())
}

func (m *FilterModel) Focus() tea.Cmd {
	m.Active = true
	return m.TextInput.Focus()
}

func (m *FilterModel) Blur() {
	m.Active = false
	m.TextInput.Blur()
	m.TextInput.Reset()
}

func (m FilterModel) Value() string {
	return m.TextInput.Value()
}
