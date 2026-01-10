package components

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yogirk/tgcp/internal/styles"
)

// Spinner frames - using common terminal spinner characters
var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// SpinnerModel represents a loading spinner component
type SpinnerModel struct {
	Frame      int
	Message    string
	Width      int
	Height     int
	FrameCount int
}

// NewSpinnerModel creates a new spinner component
func NewSpinnerModel(message string) SpinnerModel {
	return SpinnerModel{
		Message:    message,
		FrameCount: len(spinnerFrames),
	}
}

// tickMsg is sent to animate the spinner
type tickMsg time.Time

// tick returns a command that sends a tick message
func tick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Update handles messages to animate the spinner
func (m SpinnerModel) Update(msg tea.Msg) (SpinnerModel, tea.Cmd) {
	switch msg.(type) {
	case tickMsg:
		m.Frame = (m.Frame + 1) % m.FrameCount
		return m, tick()
	}
	return m, nil
}

// View renders the spinner
func (m SpinnerModel) View() string {
	spinnerChar := spinnerFrames[m.Frame]
	spinner := lipgloss.NewStyle().Foreground(styles.ColorPrimary).Render(spinnerChar)
	
	// Build the message
	var content string
	if m.Message != "" {
		content = fmt.Sprintf("%s %s", spinner, m.Message)
	} else {
		content = spinner
	}
	
	// Center the content if we have dimensions
	if m.Width > 0 && m.Height > 0 {
		return lipgloss.Place(
			m.Width, m.Height,
			lipgloss.Center, lipgloss.Center,
			content,
		)
	}
	
	return content
}

// SetSize updates the component size
func (m *SpinnerModel) SetSize(width, height int) {
	m.Width = width
	m.Height = height
}

// Init starts the spinner animation
func (m SpinnerModel) Init() tea.Cmd {
	return tick()
}

// RenderSpinner is a convenience function for services to render a loading spinner
// This is a static version that doesn't animate (for simple use cases)
func RenderSpinner(message string) string {
	if message == "" {
		message = "Loading..."
	}
	spinnerChar := spinnerFrames[0] // Use first frame for static version
	spinner := lipgloss.NewStyle().Foreground(styles.ColorPrimary).Render(spinnerChar)
	return fmt.Sprintf("%s %s", spinner, message)
}

// RenderSpinnerWithModel creates an animated spinner model
// Services should use this when they want animated spinners
func RenderSpinnerWithModel(message string) (SpinnerModel, tea.Cmd) {
	model := NewSpinnerModel(message)
	return model, model.Init()
}
