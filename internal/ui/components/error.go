package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yogirk/tgcp/internal/styles"
)

// ErrorModel represents an error display component
type ErrorModel struct {
	Error       error
	Title       string   // e.g., "Error Loading Instances"
	ServiceName string   // e.g., "GCE"
	Suggestions []string // Helpful suggestions
	Width       int
	Height      int
}

// NewErrorModel creates a new error component
func NewErrorModel(err error, title, serviceName string) ErrorModel {
	return ErrorModel{
		Error:       err,
		Title:       title,
		ServiceName: serviceName,
		Suggestions: generateSuggestions(err, serviceName),
	}
}

// Update handles messages (for future: retry button, etc.)
func (m ErrorModel) Update(msg tea.Msg) (ErrorModel, tea.Cmd) {
	// For now, error component is static
	// Future: Add retry button, dismiss button, etc.
	return m, nil
}

// View renders the error component
func (m ErrorModel) View() string {
	if m.Error == nil {
		return ""
	}

	// Error icon and title
	header := lipgloss.JoinHorizontal(
		lipgloss.Left,
		styles.ErrorStyle.Render("⚠ "),
		styles.ErrorStyle.Bold(true).Render(m.Title),
	)

	// Error message
	errorMsg := m.Error.Error()
	// Wrap long error messages
	if m.Width > 0 {
		maxWidth := m.Width - 10
		if len(errorMsg) > maxWidth {
			errorMsg = errorMsg[:maxWidth-3] + "..."
		}
	}

	// Suggestions section
	var suggestions string
	if len(m.Suggestions) > 0 {
		suggestionLines := make([]string, len(m.Suggestions))
		for i, suggestion := range m.Suggestions {
			suggestionLines[i] = fmt.Sprintf("  • %s", suggestion)
		}
		suggestions = lipgloss.JoinVertical(
			lipgloss.Left,
			styles.SubtleStyle.Render("💡 Suggestions:"),
			strings.Join(suggestionLines, "\n"),
		)
	}

	// Help text
	helpText := styles.HelpStyle.Render("[r] Retry  [q] Back")

	// Combine all parts
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		styles.ValueStyle.Render("Failed: "+errorMsg),
		"",
		suggestions,
		"",
		helpText,
	)

	// Wrap in styled box
	box := styles.BoxStyle.Copy().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorError).
		Padding(1, 2).
		Width(80).
		Render(content)

	return box
}

// generateSuggestions creates helpful suggestions based on error type
func generateSuggestions(err error, serviceName string) []string {
	errStr := err.Error()
	suggestions := []string{}

	// Permission errors
	if strings.Contains(errStr, "403") ||
		strings.Contains(errStr, "permission") ||
		strings.Contains(errStr, "Insufficient") ||
		strings.Contains(errStr, "denied") {
		suggestions = append(suggestions,
			fmt.Sprintf("Check IAM permissions for %s", serviceName),
			"Verify your account has the required roles",
		)
	}

	// Network errors
	if strings.Contains(errStr, "network") ||
		strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "connection") ||
		strings.Contains(errStr, "dial") {
		suggestions = append(suggestions,
			"Check your internet connection",
			"Verify GCP API is accessible",
		)
	}

	// Not found errors
	if strings.Contains(errStr, "404") ||
		strings.Contains(errStr, "not found") ||
		strings.Contains(errStr, "does not exist") {
		suggestions = append(suggestions,
			"Verify the resource exists",
			"Check project ID is correct",
		)
	}

	// Rate limit errors
	if strings.Contains(errStr, "429") ||
		strings.Contains(errStr, "rate limit") ||
		strings.Contains(errStr, "quota") {
		suggestions = append(suggestions,
			"API rate limit reached",
			"Wait a moment and try again",
		)
	}

	// Authentication errors
	if strings.Contains(errStr, "401") ||
		strings.Contains(errStr, "unauthorized") ||
		strings.Contains(errStr, "credentials") ||
		strings.Contains(errStr, "authentication") {
		suggestions = append(suggestions,
			"Check your GCP credentials",
			"Run: gcloud auth application-default login",
		)
	}

	// Generic fallback
	if len(suggestions) == 0 {
		suggestions = append(suggestions,
			"Try refreshing with 'r'",
			"Check project configuration",
		)
	}

	return suggestions
}

// SetSize updates the component size
func (m *ErrorModel) SetSize(width, height int) {
	m.Width = width
	m.Height = height
}

// RenderError is a convenience function for services to render errors
func RenderError(err error, serviceName, resourceType string) string {
	title := fmt.Sprintf("Error Loading %s", resourceType)
	model := NewErrorModel(err, title, serviceName)
	return model.View()
}
