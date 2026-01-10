package components

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yogirk/tgcp/internal/styles"
)

// ConfirmationModel represents a confirmation dialog component
type ConfirmationModel struct {
	Action        string // e.g., "start", "stop", "delete"
	ResourceName  string // e.g., "prod-web-1"
	ResourceType  string // e.g., "instance", "disk", "service"
	Message       string // Optional custom message (overrides default)
	Width         int
	Height        int
}

// NewConfirmationModel creates a new confirmation dialog
func NewConfirmationModel(action, resourceName, resourceType string) ConfirmationModel {
	return ConfirmationModel{
		Action:       action,
		ResourceName: resourceName,
		ResourceType: resourceType,
	}
}

// NewConfirmationModelWithMessage creates a confirmation with custom message
func NewConfirmationModelWithMessage(action, resourceName, resourceType, message string) ConfirmationModel {
	return ConfirmationModel{
		Action:       action,
		ResourceName: resourceName,
		ResourceType: resourceType,
		Message:      message,
	}
}

// Update handles messages (for future: interactive buttons, etc.)
func (m ConfirmationModel) Update(msg tea.Msg) (ConfirmationModel, tea.Cmd) {
	// For now, confirmation component is static
	// Services handle keybindings themselves
	return m, nil
}

// View renders the confirmation dialog
func (m ConfirmationModel) View() string {
	// Title
	title := styles.WarningStyle.Bold(true).Render("⚠ Confirm Action")

	// Action text
	var actionText string
	if m.Message != "" {
		actionText = m.Message
	} else {
		actionText = m.buildActionText()
	}

	// Help text
	helpText := styles.HelpStyle.Render("y (Confirm) / n (Cancel)")

	// Combine content
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		"",
		actionText,
		"",
		helpText,
	)

	// Wrap in styled box
	dialog := styles.BoxStyle.Copy().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorWarning).
		Padding(1, 4).
		Width(70).
		Render(content)

	// Center the dialog
	return lipgloss.Place(
		80, 20, // Approximate dimensions
		lipgloss.Center, lipgloss.Center,
		dialog,
	)
}

// buildActionText constructs the confirmation message based on action type
func (m ConfirmationModel) buildActionText() string {
	actionUpper := capitalize(m.Action)
	resourceNameStyled := styles.TitleStyle.Render(m.ResourceName)
	
	// Build action verb based on action type
	var verb string
	switch m.Action {
	case "start":
		verb = "START"
	case "stop":
		verb = "STOP"
	case "delete":
		verb = "DELETE"
	case "restart":
		verb = "RESTART"
	case "snapshot":
		verb = "CREATE SNAPSHOT OF"
	default:
		verb = actionUpper
	}

	return fmt.Sprintf(
		"Are you sure you want to %s %s %s?",
		verb,
		m.ResourceType,
		resourceNameStyled,
	)
}

// capitalize capitalizes the first letter of a string
func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	// Convert first character to uppercase
	first := s[0]
	if first >= 'a' && first <= 'z' {
		first = first - 32
	}
	return string(first) + s[1:]
}

// SetSize updates the component size
func (m *ConfirmationModel) SetSize(width, height int) {
	m.Width = width
	m.Height = height
}

// RenderConfirmation is a convenience function for services to render confirmation dialogs
func RenderConfirmation(action, resourceName, resourceType string) string {
	model := NewConfirmationModel(action, resourceName, resourceType)
	return model.View()
}

// RenderConfirmationWithMessage renders a confirmation with a custom message
func RenderConfirmationWithMessage(action, resourceName, resourceType, message string) string {
	model := NewConfirmationModelWithMessage(action, resourceName, resourceType, message)
	return model.View()
}
