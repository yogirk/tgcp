package components

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yogirk/tgcp/internal/styles"
)

// SpinnerTickMsg is sent to animate the spinner
type SpinnerTickMsg time.Time

// Glowing star frames - creates a pulsing effect
var starFrames = []string{"✦", "✧", "⋆", "✧"}

// Color cycle for the glow effect (dim -> bright -> white -> bright -> dim)
var glowColors = []lipgloss.Color{
	lipgloss.Color("33"),  // dim blue
	lipgloss.Color("39"),  // brand blue
	lipgloss.Color("75"),  // bright blue
	lipgloss.Color("117"), // very bright
	lipgloss.Color("231"), // white (peak)
	lipgloss.Color("117"), // very bright
	lipgloss.Color("75"),  // bright blue
	lipgloss.Color("39"),  // brand blue
}

// Playful loading messages
var playfulMessages = []string{
	"Consulting the cloud...",
	"Waking up the instances...",
	"Counting your resources...",
	"Asking GCP nicely...",
	"Fetching the good stuff...",
	"Poking the APIs...",
	"Loading cloud magic...",
	"Almost there...",
	"Gathering data...",
	"Doing cloud things...",
}

// SpinnerModel represents an animated loading spinner
type SpinnerModel struct {
	frame        int       // Current animation frame
	messageIndex int       // Current playful message
	message      string    // Custom message (overrides playful if set)
	lastMsgSwap  time.Time // When we last changed the playful message
	active       bool      // Whether spinner is active
}

// NewSpinner creates a new animated spinner
func NewSpinner() SpinnerModel {
	return SpinnerModel{
		lastMsgSwap: time.Now(),
	}
}

// Start activates the spinner with an optional custom message
func (m *SpinnerModel) Start(customMessage string) tea.Cmd {
	m.active = true
	m.message = customMessage
	m.frame = 0
	m.messageIndex = 0
	m.lastMsgSwap = time.Now()
	return m.tick()
}

// Stop deactivates the spinner
func (m *SpinnerModel) Stop() {
	m.active = false
}

// IsActive returns whether the spinner is currently active
func (m *SpinnerModel) IsActive() bool {
	return m.active
}

// tick returns a command that sends a tick message
func (m SpinnerModel) tick() tea.Cmd {
	return tea.Tick(80*time.Millisecond, func(t time.Time) tea.Msg {
		return SpinnerTickMsg(t)
	})
}

// Update handles animation ticks
func (m SpinnerModel) Update(msg tea.Msg) (SpinnerModel, tea.Cmd) {
	if !m.active {
		return m, nil
	}

	switch msg.(type) {
	case SpinnerTickMsg:
		// Advance frame
		m.frame = (m.frame + 1) % len(glowColors)

		// Rotate playful message every ~2 seconds
		if m.message == "" && time.Since(m.lastMsgSwap) > 2*time.Second {
			m.messageIndex = (m.messageIndex + 1) % len(playfulMessages)
			m.lastMsgSwap = time.Now()
		}

		return m, m.tick()
	}
	return m, nil
}

// View renders the animated spinner as a simple inline line
func (m SpinnerModel) View() string {
	if !m.active {
		return ""
	}

	// Get current star character (cycles slower than color)
	starIdx := (m.frame / 2) % len(starFrames)
	star := starFrames[starIdx]

	// Get current glow color
	color := glowColors[m.frame]

	// Style the star with current glow color
	starStyle := lipgloss.NewStyle().
		Foreground(color).
		Bold(true)

	// Get message
	msg := m.message
	if msg == "" {
		msg = playfulMessages[m.messageIndex]
	}

	// Message style
	msgStyle := lipgloss.NewStyle().
		Foreground(styles.ColorTextMuted).
		Italic(true)

	// Simple inline: ✦ Loading message...
	return "  " + starStyle.Render(star) + " " + msgStyle.Render(msg)
}

// ViewCentered renders the spinner centered in the given dimensions
func (m SpinnerModel) ViewCentered(width, height int) string {
	if !m.active {
		return ""
	}

	return lipgloss.Place(
		width, height,
		lipgloss.Center, lipgloss.Center,
		m.View(),
	)
}

// --- Legacy compatibility functions ---
// These are kept for backwards compatibility but now just return static text.
// Services should transition to using the centralized spinner in MainModel.

// RenderSpinner returns a static loading indicator (deprecated - use MainModel spinner)
func RenderSpinner(message string) string {
	if message == "" {
		message = "Loading..."
	}
	star := lipgloss.NewStyle().
		Foreground(styles.ColorBrandPrimary).
		Bold(true).
		Render("✦")
	msg := lipgloss.NewStyle().
		Foreground(styles.ColorTextMuted).
		Italic(true).
		Render(message)
	return star + " " + msg
}
