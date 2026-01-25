package logging

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yogirk/tgcp/internal/core"
	"github.com/yogirk/tgcp/internal/styles"
	"github.com/yogirk/tgcp/internal/ui/components"
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
	filter  components.FilterModel
	spinner components.SpinnerModel

	// State
	entries       []LogEntry
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
	// Viewport Setup
	vp := viewport.New(0, 0)
	vp.Style = lipgloss.NewStyle().
		PaddingRight(0)

	s := &Service{
		viewport: vp,
		filter:   components.NewFilterWithPlaceholder("Filter (e.g. severity>=ERROR)"),
		spinner:  components.NewSpinner(),
		cache:    cache,
	}
	s.filter.TextInput.CharLimit = 1000 // Allow long filters
	return s
}

func (s *Service) Name() string {
	return "Cloud Logging"
}

func (s *Service) ShortName() string {
	return "logs"
}

func (s *Service) HelpText() string {
	if s.filter.IsActive() {
		return "Esc:Cancel  Enter:Apply"
	}
	base := "r:Refresh  /:Filter  Esc/q:Back"
	if s.returnTo != "" {
		base = fmt.Sprintf("Esc:Back to %s  r:Refresh  /:Filter", s.returnTo)
	}
	if len(s.tokenStack) > 0 {
		base += "  n:Newer"
	}
	if s.nextPageToken != "" {
		base += "  p:Older"
	}
	return base
}

// Focus handles input focus
func (s *Service) Focus() {
	// Viewport always active
}

// Blur handles loss of input focus
func (s *Service) Blur() {
	s.filter.TextInput.Blur()
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
	return nil
}

// Reinit reinitializes the service with a new project ID
func (s *Service) Reinit(ctx context.Context, projectID string) error {
	s.Reset()
	return s.InitService(ctx, projectID)
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
	case components.SpinnerTickMsg:
		s.spinner, cmd = s.spinner.Update(msg)
		return s, cmd

	case tickMsg:
		// Background refresh always fetches page 1 (empty token) to see latest
		return s, tea.Batch(s.fetchEntriesCmd(""), s.tick())

	case entriesMsg:
		s.spinner.Stop()
		s.entries = msg.entries
		s.nextPageToken = msg.nextToken
		// Current token is already set before fetch
		s.renderLogs()
		return s, func() tea.Msg { return core.LastUpdatedMsg(time.Now()) }

	case errMsg:
		s.spinner.Stop()
		s.err = msg

	case tea.WindowSizeMsg:
		const heightOffset = 6
		newHeight := msg.Height - heightOffset
		if newHeight < 5 {
			newHeight = 5
		}
		s.viewport.Width = msg.Width
		s.viewport.Height = newHeight
		s.filter.TextInput.Width = msg.Width - 4 // Dynamic filter width
		s.renderLogs()                           // Re-render triggers wrapping

	case tea.KeyMsg:
		// FILTERING MODE delegated to component
		// Only check global keys if NOT filtering
		if !s.filter.IsActive() {
			switch msg.String() {
			case "/":
				return s, s.filter.EnterFilterMode()
			case "r":
				return s, s.Refresh()

			case "p": // Previous / Older (Fetch next API page)
				if s.nextPageToken != "" {
					s.spinner.Start("")
					// Push current token to stack (which represents state of "newer" page)
					s.tokenStack = append(s.tokenStack, s.currentToken)
					// Update current token
					s.currentToken = s.nextPageToken
					return s, s.fetchEntriesCmd(s.currentToken)
				}
			case "n": // Next / Newer (Pop stack)
				if len(s.tokenStack) > 0 {
					s.spinner.Start("")
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
				// Default behavior for Help bubbling?
				return s, nil
			}
		}

		// Handle Filter Input
		var filterCmd tea.Cmd
		// We don't use FilterSession here because we have custom "Apply" logic (Refresh)
		// So we handle filter keys manually or use the component's Update
		s.filter, filterCmd = s.filter.Update(msg)
		if s.filter.IsActive() {
			// Check for Enter/Esc which might be handled by component or us
			// Actually component handles value update, we check for Enter to trigger fetch
			if msg.String() == "enter" {
				s.filter.ExitFilterMode()
				return s, s.Refresh()
			}
			if msg.String() == "esc" {
				s.filter.ExitFilterMode()
				// Don't reset if just cancelling?
				// But we did exit match
				return s, nil
			}
			return s, filterCmd
		}

		s.viewport, cmd = s.viewport.Update(msg)
		return s, tea.Batch(cmd, filterCmd)
	}

	return s, cmd
}

// SetFilter sets the filter string (used by other components to jump to logs)
func (s *Service) SetFilter(filter string) {
	s.filter.TextInput.SetValue(filter)
	s.filter.TextInput.SetCursor(0)
	s.filter.ExitFilterModeKeepValue()
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
	msgStyle := lipgloss.NewStyle().Foreground(styles.ColorTextPrimary)

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

		// Use Value() from TextInput in FilterModel
		filter := s.filter.TextInput.Value()
		pageSize := 12 // Requested constraint

		entries, nextToken, err := s.client.ListEntries(context.Background(), filter, pageSize, token)
		if err != nil {
			return errMsg(err)
		}
		return entriesMsg{entries: entries, nextToken: nextToken}
	}
}

func (s *Service) Refresh() tea.Cmd {
	s.spinner.Start("")
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
