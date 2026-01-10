package components

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yogirk/tgcp/internal/styles"
)

type StatusMsg string

type StatusBarModel struct {
	Message     string
	Mode        string // "NORMAL", "COMMAND", "FILTER"
	FocusPane   string // "HOME", "SIDEBAR", "MAIN"
	HelpText    string
	Width       int
	LastUpdated time.Time
	IsError     bool
}

func NewStatusBar() StatusBarModel {
	return StatusBarModel{
		Message:  "Ready",
		Mode:     "NORMAL",
		FocusPane: "",
		HelpText: "", // Dynamically set by view
		Width:    80,
	}
}

// SetHelpText updates the help text
func (m *StatusBarModel) SetHelpText(text string) {
	m.HelpText = text
}

// SetFocusPane updates the active pane indicator
func (m *StatusBarModel) SetFocusPane(pane string) {
	m.FocusPane = pane
}

func (m StatusBarModel) Init() tea.Cmd {
	return nil
}

func (m StatusBarModel) Update(msg tea.Msg) (StatusBarModel, tea.Cmd) {
	switch msg := msg.(type) {
	case StatusMsg:
		m.Message = string(msg)
	}
	return m, nil
}

func (m StatusBarModel) View() string {
	// Styles
	modeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("232")).
		Background(styles.ColorSecondary). // Default
		Bold(true).
		Padding(0, 1)

	modeLabel := m.Mode
	if m.IsError {
		modeStyle = modeStyle.Background(lipgloss.Color("196")) // Red
		modeLabel = "ERROR"
	} else if m.Mode == "COMMAND" {
		modeStyle = modeStyle.Background(styles.ColorPrimary) // Pink
	} else if m.Mode == "FILTER" {
		modeStyle = modeStyle.Background(styles.ColorWarning) // Orange
	} else if m.FocusPane == "MAIN" {
		modeStyle = modeStyle.Background(styles.ColorAccent)
		modeLabel = m.FocusPane
	} else if m.FocusPane == "SIDEBAR" || m.FocusPane == "HOME" {
		modeStyle = modeStyle.Background(styles.ColorSecondary)
		modeLabel = m.FocusPane
	} else if m.Mode == "NORMAL" {
		modeLabel = "NORMAL"
	}

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Background(lipgloss.Color("237")).
		Align(lipgloss.Right).
		Padding(0, 1)

	// Layout
	// [ MODE ] [ Message .......................... ] [ Help ]

	mode := modeStyle.Render(modeLabel)

	// Right side: Last Updated + Help
	var rightSide string
	if !m.LastUpdated.IsZero() {
		since := time.Since(m.LastUpdated).Round(time.Second)
		timeStr := fmt.Sprintf("Updated: %s ago", since)
		rightSide = styles.SubtleStyle.Render(timeStr) + "  " + helpStyle.Render(m.HelpText)
	} else {
		rightSide = helpStyle.Render(m.HelpText)
	}

	// Calculate available width for message
	infoWidth := m.Width - lipgloss.Width(mode) - lipgloss.Width(rightSide)
	if infoWidth < 0 {
		infoWidth = 0
	}

	info := styles.StatusBarStyle.Width(infoWidth).Render(m.Message)

	return lipgloss.JoinHorizontal(lipgloss.Top, mode, info, rightSide)
}
