package logging

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/yogirk/tgcp/internal/core"
	"github.com/yogirk/tgcp/internal/ui/components"
)

const CacheTTL = 10 * time.Second // Logs change frequently

// Tick message for background refresh
type tickMsg time.Time

// Service implements the services.Service interface for Cloud Logging
type Service struct {
	client    *Client
	projectID string

	// UI Components
	table   *components.StandardTable
	spinner components.SpinnerModel

	// State
	entries       []LogEntry
	err           error
	nextPageToken string
	currentToken  string   // Token used for current page
	tokenStack    []string // History of tokens for "Previous" function

	// Dimensions
	width  int
	height int

	// Cache
	cache *core.Cache

	// Navigation
	returnTo string
	heading  string

	// Filter for API calls
	filter string

	// Detail view
	selectedEntry *LogEntry
	viewingDetail bool
}

func NewService(cache *core.Cache) *Service {
	// Compact table columns for log entries
	columns := []table.Column{
		{Title: "Time", Width: 19},
		{Title: "Sev", Width: 8},
		{Title: "Message", Width: 80},
	}

	t := components.NewStandardTable(columns)

	s := &Service{
		table:   t,
		spinner: components.NewSpinner(),
		cache:   cache,
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
	if s.viewingDetail {
		return "Esc/q:Back"
	}
	base := "r:Refresh  Enter:Detail"
	if s.returnTo != "" {
		base = fmt.Sprintf("Esc:Back to %s  r:Refresh  Enter:Detail", s.returnTo)
	} else {
		base += "  Esc/q:Back"
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
	s.table.Focus()
}

// Blur handles loss of input focus
func (s *Service) Blur() {
	s.table.Blur()
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
		s.updateTable()
		return s, func() tea.Msg { return core.LastUpdatedMsg(time.Now()) }

	case errMsg:
		s.spinner.Stop()
		s.err = msg

	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
		s.table.HandleWindowSizeDefault(msg)
		// Adjust message column to fill available width
		s.adjustTableColumns()

	case tea.KeyMsg:
		if s.viewingDetail {
			switch msg.String() {
			case "esc", "q":
				s.viewingDetail = false
				s.selectedEntry = nil
				return s, nil
			}
			return s, nil
		}

		switch msg.String() {
		case "r":
			return s, s.Refresh()

		case "enter":
			if idx := s.table.Cursor(); idx >= 0 && idx < len(s.entries) {
				s.selectedEntry = &s.entries[idx]
				s.viewingDetail = true
			}
			return s, nil

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
			return s, nil
		}

		// Pass to table for navigation
		var updatedTable *components.StandardTable
		updatedTable, cmd = s.table.Update(msg)
		s.table = updatedTable
		return s, cmd
	}

	return s, cmd
}

// SetFilter sets the filter string (used by other components to jump to logs)
func (s *Service) SetFilter(filter string) {
	s.filter = filter
}

// SetReturnTo sets the service to return to when Esc is pressed
func (s *Service) SetReturnTo(service string) {
	s.returnTo = service
}

// SetHeading sets the custom heading
func (s *Service) SetHeading(heading string) {
	s.heading = heading
}

func (s *Service) adjustTableColumns() {
	// Calculate available width for message column
	// Time=19, Sev=8, padding/borders ~6
	msgWidth := s.width - 19 - 8 - 10
	if msgWidth < 30 {
		msgWidth = 30
	}
	if msgWidth > 120 {
		msgWidth = 120
	}

	columns := []table.Column{
		{Title: "Time", Width: 19},
		{Title: "Sev", Width: 8},
		{Title: "Message", Width: msgWidth},
	}
	s.table.SetColumns(columns)
}

func (s *Service) updateTable() {
	rows := make([]table.Row, len(s.entries))
	for i, e := range s.entries {
		ts := e.Timestamp.Local().Format("01-02 15:04:05")
		sev := formatSeverityShort(e.Severity)

		// Truncate message for table display
		msg := e.Payload
		// Remove newlines for table display
		msg = strings.ReplaceAll(msg, "\n", " ")
		if len(msg) > 100 {
			msg = msg[:97] + "..."
		}

		rows[i] = table.Row{ts, sev, msg}
	}
	s.table.SetRows(rows)
}

func formatSeverityShort(severity string) string {
	switch severity {
	case "EMERGENCY":
		return components.RenderStatus("EMERG")
	case "CRITICAL":
		return components.RenderStatus("CRIT")
	case "ERROR":
		return components.RenderStatus("ERROR")
	case "WARNING":
		return components.RenderStatus("WARN")
	case "NOTICE":
		return components.RenderStatus("NOTICE")
	case "INFO":
		return components.RenderStatus("INFO")
	case "DEBUG":
		return components.RenderStatus("DEBUG")
	default:
		return components.RenderStatus("DEFAULT")
	}
}

func (s *Service) fetchEntriesCmd(token string) tea.Cmd {
	return func() tea.Msg {
		if s.client == nil {
			return errMsg(fmt.Errorf("client not initialized"))
		}

		// Use the stored filter value
		filter := s.filter
		pageSize := 25 // More entries fit now with compact table format

		entries, nextToken, err := s.client.ListEntries(context.Background(), filter, pageSize, token)
		if err != nil {
			return errMsg(err)
		}
		return entriesMsg{entries: entries, nextToken: nextToken}
	}
}

func (s *Service) Refresh() tea.Cmd {
	s.spinner.Start("")
	s.currentToken = ""
	s.tokenStack = nil
	return s.fetchEntriesCmd("")
}

func (s *Service) Reset() {
	s.err = nil
	s.viewingDetail = false
	s.selectedEntry = nil
	s.table.SetCursor(0)
	s.SetHeading("")
	s.currentToken = ""
	s.tokenStack = nil
	s.filter = "" // Clear filter on reset
}

func (s *Service) IsRootView() bool {
	return !s.viewingDetail
}
