package logging

import (
	"context"
	"fmt"
	"time"

	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rk/tgcp/internal/core"
	"github.com/rk/tgcp/internal/styles"
)

const CacheTTL = 10 * time.Second // Logs change frequently

// Tick message for background refresh
type tickMsg time.Time

// Service implements the services.Service interface for Cloud Logging
type Service struct {
	client    *Client
	projectID string
	viewport  viewport.Model

	// UI Components
	filterInput textinput.Model
	filtering   bool

	// State
	entries []LogEntry
	loading       bool
	err           error
	nextPageToken string
	currentToken  string   // Token used for current page
	tokenStack    []string // History of tokens for "Previous" function

	// Cache
	cache *core.Cache

	// Navigation
	returnTo string
	heading  string
}

func NewService(cache *core.Cache) *Service {
	// Table Setup
	// Viewport Setup
	vp := viewport.New(0, 0)
	vp.Style = lipgloss.NewStyle().
		PaddingRight(0)

	// Filter Input Setup
	ti := textinput.New()
	ti.Placeholder = "Filter (e.g. severity>=ERROR)"
	ti.Prompt = "/ "
	ti.CharLimit = 1000
	ti.Width = 60 // Initial default
	ti.SetValue("") // Default to empty (triggers last 30m in api.go)

	s := &Service{
		viewport:    vp,
		filterInput: ti,
		cache:       cache,
	}
	return s
}

func (s *Service) Name() string {
	return "Cloud Logging"
}

func (s *Service) ShortName() string {
	return "logs"
}

func (s *Service) HelpText() string {
	if s.filtering {
		return "Esc:Cancel  Enter:Apply"
	}
	base := "r:Refresh  /:Filter  Esc/q:Back"
	if s.returnTo != "" {
		base = fmt.Sprintf("Esc:Back to %s  r:Refresh  /:Filter", s.returnTo)
	}
	if s.nextPageToken != "" {
		base += "  n:Next"
	}
	if len(s.tokenStack) > 0 {
		base += "  p:Prev"
	}
	return base
}



// Focus handles input focus
func (s *Service) Focus() {
	// table was focused, nothing to do for viewport which always accepts keys if forwarded

}

// Blur handles loss of input focus
func (s *Service) Blur() {
	// table blur logic removed

}

// Msg types
type entriesMsg struct {
	entries   []LogEntry
	nextToken string
}
type errMsg error

// InitService initializes the service logic (API clients)
func (s *Service) InitService(ctx context.Context, projectID string) error {
	s.projectID = projectID
	client, err := NewClient(ctx, projectID)
	if err != nil {
		return err
	}
	s.client = client
	
	// Initial fetch
	// We return nil here as this is synchronous init called by app
	// The actual fetch happens via Refresh/Init command usually
	return nil
}

// Init satisfies tea.Model interface, starts background tick
func (s *Service) Init() tea.Cmd {
	return tea.Batch(
		s.tick(),
		s.Refresh(), // Trigger first fetch
	)
}

func (s *Service) tick() tea.Cmd {
	return tea.Tick(CacheTTL, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Update handles messages specific to Logging
func (s *Service) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tickMsg:
		// Background refresh always fetches page 1 (empty token) to see latest
		return s, tea.Batch(s.fetchEntriesCmd(""), s.tick())

	case entriesMsg:
		s.loading = false
		s.entries = msg.entries
		s.nextPageToken = msg.nextToken
		// Current token is already set before fetch
		s.renderLogs()
		return s, func() tea.Msg { return core.LastUpdatedMsg(time.Now()) }

	case errMsg:
		s.loading = false
		s.err = msg

	case tea.WindowSizeMsg:
		const heightOffset = 6
		newHeight := msg.Height - heightOffset
		if newHeight < 5 {
			newHeight = 5
		}
		s.viewport.Width = msg.Width
		s.viewport.Height = newHeight
		s.filterInput.Width = msg.Width - 4 // Dynamic filter width
		s.renderLogs() // Re-render triggers wrapping


	case tea.KeyMsg:
		// FILTERING MODE
		if s.filtering {
			switch msg.String() {
			case "esc":
				if s.returnTo != "" {
					dest := s.returnTo
					s.returnTo = "" // Reset
					return s, func() tea.Msg { return core.SwitchToServiceMsg{Service: dest} }
				}
				s.filtering = false
				s.filterInput.Blur()
				// Revert to valid filter or keep?
				// s.filterInput.Reset() 
				return s, nil
			case "enter":
				s.filtering = false
				s.filterInput.Blur()
				return s, s.Refresh() // Apply filter
			}

			var inputCmd tea.Cmd
			s.filterInput, inputCmd = s.filterInput.Update(msg)
			return s, inputCmd
		}

		// LIST VIEW
		switch msg.String() {
		case "/":
			s.filtering = true
			s.filterInput.Focus()
			return s, textinput.Blink
		case "r":
			return s, s.Refresh()
		case "n":
			if s.nextPageToken != "" {
				s.loading = true
				// Push current token to stack
				s.tokenStack = append(s.tokenStack, s.currentToken)
				// Update current token
				s.currentToken = s.nextPageToken
				return s, s.fetchEntriesCmd(s.currentToken)
			}
		case "p":
			if len(s.tokenStack) > 0 {
				s.loading = true
				// Pop last token
				lastIdx := len(s.tokenStack) - 1
				prevToken := s.tokenStack[lastIdx]
				s.tokenStack = s.tokenStack[:lastIdx]
				
				// Update current token
				s.currentToken = prevToken
				return s, s.fetchEntriesCmd(s.currentToken)
			}
		case "esc", "q":
			if s.returnTo != "" {
				dest := s.returnTo
				s.returnTo = "" // Reset
				return s, func() tea.Msg { return core.SwitchToServiceMsg{Service: dest} }
			}
			return s, nil // Let parent handle if no returnTo? Or just Consume?
		}		
		s.viewport, cmd = s.viewport.Update(msg)
		return s, cmd
	}

	return s, cmd
}

// SetFilter sets the filter string (used by other components to jump to logs)
func (s *Service) SetFilter(filter string) {
	s.filterInput.SetValue(filter)
	s.filterInput.SetCursor(0) // Scroll to start so user sees the beginning context
	s.filtering = false // Just apply the filter, don't enter edit mode
	// If we want to auto-apply, we might need to trigger refresh, but this is just setting state.
	// The caller (SwitchToLogsMsg) calls Refresh() right after.
}

// SetReturnTo sets the service to return to when Esc is pressed
func (s *Service) SetReturnTo(service string) {
	s.returnTo = service
}

// SetHeading sets the custom heading and adjusts columns
func (s *Service) SetHeading(heading string) {
	s.heading = heading
	// Re-render handled by next update/fetch or we can force it if we had entries
	if len(s.entries) > 0 {
		s.renderLogs()
	}
}

func (s *Service) renderLogs() {
	doc := strings.Builder{}

	// Define Styles
	metaStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")) // Grey for timestamp/location
	msgStyle := lipgloss.NewStyle().Foreground(styles.ColorText)

	// Determine effective width for wrapping
	// Viewport width might be 0 initially
	wrapWidth := s.viewport.Width - 4
	if wrapWidth < 20 {
		wrapWidth = 60 // fallback
	}

	for _, e := range s.entries {
		ts := e.Timestamp.Local().Format("2006-01-02 15:04:05")
		sev := renderSeverity(e.Severity)

		// Header Line: TIMESTAMP  SEVERITY  LOCATION
		// Combine location info
		loc := e.Location
		if s.heading == "" {
			if e.ResourceType != "" {
				loc = fmt.Sprintf("%s %s %s", e.ResourceType, e.ResourceName, e.Location)
			}
		}

		header := fmt.Sprintf("%s  %s  %s", ts, sev, loc)
		doc.WriteString(metaStyle.Render(header) + "\n")

		// Message Body (Wrapped)
		// We use lipgloss to wrap the payload to the viewport width
		payload := e.Payload
		wrapped := msgStyle.Width(wrapWidth).Render(payload)
		
		doc.WriteString(wrapped + "\n")
		
		// Separator
		doc.WriteString(metaStyle.Render(strings.Repeat("-", wrapWidth)) + "\n")
	}

	s.viewport.SetContent(doc.String())
}


func (s *Service) fetchEntriesCmd(token string) tea.Cmd {
	return func() tea.Msg {
		if s.client == nil {
			return errMsg(fmt.Errorf("client not initialized"))
		}

		filter := s.filterInput.Value()
		pageSize := 12 // Requested constraint
		
		entries, nextToken, err := s.client.ListEntries(context.Background(), filter, pageSize, token)
		if err != nil {
			return errMsg(err)
		}
		return entriesMsg{entries: entries, nextToken: nextToken}
	}
}

func (s *Service) Refresh() tea.Cmd {
	s.loading = true
	// Refresh keeps current page? Or resets to top?
	// Usually Refresh means "give me latest", which effectively means page 1
	// If user wants to reload current page, that's different.
	// We'll reset to page 1 for standard "Refresh" behavior
	s.currentToken = ""
	s.tokenStack = nil
	return s.fetchEntriesCmd("")
}

func (s *Service) Reset() {
	s.err = nil
	s.viewport.GotoTop()
	s.SetHeading("")
	s.currentToken = ""
	s.tokenStack = nil
}

func (s *Service) IsRootView() bool {
	return true
}

// updateTable removed

