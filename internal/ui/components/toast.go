package components

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yogirk/tgcp/internal/core"
	"github.com/yogirk/tgcp/internal/styles"
)

// ToastDismissMsg is sent when a toast should be dismissed
type ToastDismissMsg struct{}

// ToastModel represents a temporary notification
type ToastModel struct {
	Message   string
	Type      core.ToastType
	Duration  time.Duration
	CreatedAt time.Time
}

// NewToast creates a new toast notification
func NewToast(message string, toastType core.ToastType, duration time.Duration) *ToastModel {
	return &ToastModel{
		Message:   message,
		Type:      toastType,
		Duration:  duration,
		CreatedAt: time.Now(),
	}
}

// NewToastFromMsg creates a toast from a ToastMsg
func NewToastFromMsg(msg core.ToastMsg) *ToastModel {
	duration := msg.Duration
	if duration == 0 {
		duration = 3 * time.Second // default
	}
	return NewToast(msg.Message, msg.Type, duration)
}

// IsExpired checks if the toast has expired
func (t *ToastModel) IsExpired() bool {
	return time.Since(t.CreatedAt) > t.Duration
}

// DismissCmd returns a command that will dismiss the toast after duration
func (t *ToastModel) DismissCmd() tea.Cmd {
	return tea.Tick(t.Duration, func(time.Time) tea.Msg {
		return ToastDismissMsg{}
	})
}

// View renders the toast notification
func (t *ToastModel) View() string {
	if t == nil {
		return ""
	}

	var icon string
	var borderColor lipgloss.Color

	switch t.Type {
	case core.ToastSuccess:
		icon = "✓"
		borderColor = styles.ColorSuccess
	case core.ToastError:
		icon = "✗"
		borderColor = styles.ColorError
	case core.ToastInfo:
		icon = "ℹ"
		borderColor = styles.ColorBrandPrimary
	}

	// Toast style
	toastStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 2).
		Background(lipgloss.Color("235"))

	// Icon style
	iconStyle := lipgloss.NewStyle().
		Foreground(borderColor).
		Bold(true)

	// Message style
	msgStyle := lipgloss.NewStyle().
		Foreground(styles.ColorTextPrimary)

	content := iconStyle.Render(icon) + " " + msgStyle.Render(t.Message)

	return toastStyle.Render(content)
}

