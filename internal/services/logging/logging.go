package logging

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
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
	table     table.Model

	// UI Components
	filterInput textinput.Model
	filtering   bool

	// State
	entries []LogEntry
	loading       bool
	err           error
	nextPageToken string

	// Cache
	cache *core.Cache

	// Navigation
	returnTo string
	heading  string
}

func NewService(cache *core.Cache) *Service {
	// Table Setup
	t := table.New(
		table.WithFocused(true),
		table.WithHeight(10),
	)

	// Custom Table Styles
	st := table.DefaultStyles()
	st.Header = styles.HeaderStyle
	st.Selected = lipgloss.NewStyle().
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	
	// Add bottom border to all cells for "clear separated line"
	st.Cell = st.Cell.Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1) // Add horizontal padding for "clear indentation"

	t.SetStyles(st)

	// Filter Input Setup
	ti := textinput.New()
	ti.Placeholder = "Filter (e.g. severity>=ERROR)"
	ti.Prompt = "/ "
	ti.CharLimit = 200
	ti.Width = 60
	ti.SetValue("") // Default to empty (triggers last 30m in api.go)

	s := &Service{
		table:       t,
		filterInput: ti,
		cache:       cache,
	}
	s.setColumns(false) // Default columns
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
		base += "  n:Next Page"
	}
	return base
}



// Focus handles input focus
func (s *Service) Focus() {
	s.table.Focus()
	st := table.DefaultStyles()
	st.Header = styles.HeaderStyle
	st.Selected = lipgloss.NewStyle().
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	s.table.SetStyles(st)
}

// Blur handles loss of input focus
func (s *Service) Blur() {
	s.table.Blur()
	st := table.DefaultStyles()
	st.Header = styles.HeaderStyle
	st.Selected = lipgloss.NewStyle().
		Foreground(styles.ColorText).
		Background(lipgloss.Color("237")). // Dark grey
		Bold(false)
	s.table.SetStyles(st)
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
		s.updateTable(s.entries)
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
		s.table.SetHeight(newHeight)

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
				return s, s.fetchEntriesCmd(s.nextPageToken)
			}
		// case "enter": // View Details
		}
		
		s.table, cmd = s.table.Update(msg)
		return s, cmd
	}

	return s, cmd
}

// SetFilter sets the filter string (used by other components to jump to logs)
func (s *Service) SetFilter(filter string) {
	s.filterInput.SetValue(filter)
	s.filtering = true // optional: enter filtering mode or just apply?
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
	s.setColumns(heading != "")
}

func (s *Service) setColumns(minimal bool) {
	var cols []table.Column
	if minimal {
		cols = []table.Column{
			{Title: "Time", Width: 25},
			{Title: "Severity", Width: 10},
			{Title: "Location", Width: 15},
			{Title: "Payload", Width: 80}, // Expanded payload
		}
	} else {
		cols = []table.Column{
			{Title: "Time", Width: 25},
			{Title: "Severity", Width: 10},
			{Title: "Resource", Width: 20},
			{Title: "Name", Width: 20},
			{Title: "Location", Width: 15},
			{Title: "Payload", Width: 60},
		}
	}
	s.table.SetColumns(cols)
}


func (s *Service) fetchEntriesCmd(token string) tea.Cmd {
	return func() tea.Msg {
		if s.client == nil {
			return errMsg(fmt.Errorf("client not initialized"))
		}

		filter := s.filterInput.Value()
		pageSize := 20 // Requirement: 10 entries by default
		
		entries, nextToken, err := s.client.ListEntries(context.Background(), filter, pageSize, token)
		if err != nil {
			return errMsg(err)
		}
		return entriesMsg{entries: entries, nextToken: nextToken}
	}
}

func (s *Service) Refresh() tea.Cmd {
	s.loading = true
	return s.fetchEntriesCmd("")
}

func (s *Service) Reset() {
	s.err = nil
	s.table.SetCursor(0)
	s.SetHeading("") // Reset heading and columns
}

func (s *Service) IsRootView() bool {
	return true
}

func (s *Service) updateTable(entries []LogEntry) {
	rows := make([]table.Row, len(entries))
	for i, e := range entries {
		ts := e.Timestamp.Local().Format("2006-01-02 15:04:05")
		sev := renderSeverity(e.Severity)

		if s.heading != "" {
			// Minimal columns
			rows[i] = table.Row{
				ts,
				sev,
				e.Location,
				e.Payload,
			}
		} else {
			// Full columns
			rows[i] = table.Row{
				ts,
				sev,
				e.ResourceType,
				e.ResourceName,
				e.Location,
				e.Payload,
			}
		}
	}
	s.table.SetRows(rows)
}
