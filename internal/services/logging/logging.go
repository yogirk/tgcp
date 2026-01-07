package logging

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rk/tgcp/internal/core"
	"github.com/rk/tgcp/internal/styles"
	"github.com/rk/tgcp/internal/ui/components"
)

// ViewState represents the current view
type ViewState int

const (
	ViewList ViewState = iota
	ViewDetail
)

// Service implements the logging service
type Service struct {
	client    *Client
	projectID string
	cache     *core.Cache
	ctx       context.Context
	cancel    context.CancelFunc

	// State
	viewState   ViewState
	entries     []LogEntry
	selectedLog *LogEntry
	filter      string
	loading     bool
	err         error
	mu          sync.RWMutex

	// UI Components
	table    table.Model
	viewport viewport.Model
}

// NewService creates a new logging service
func NewService(cache *core.Cache) *Service {
	// Table Setup
	columns := []table.Column{
		{Title: "Time", Width: 20},
		{Title: "Severity", Width: 10},
		{Title: "Resource", Width: 15},
		{Title: "Message", Width: 50},
	}
	t := components.NewTable(columns, []table.Row{})

	return &Service{
		cache:     cache,
		table:     t.Table, // Use the underlying table.Model
		viewState: ViewList,
	}
}

// Init implements tea.Model
func (s *Service) Init() tea.Cmd {
	return nil
}

// Name returns the display name
func (s *Service) Name() string {
	return "Cloud Logging"
}

// ShortName returns the ID
func (s *Service) ShortName() string {
	return "logs"
}

// InitService initializes the service
func (s *Service) InitService(ctx context.Context, projectID string) error {
	s.projectID = projectID
	s.ctx, s.cancel = context.WithCancel(ctx)

	client, err := NewClient(s.ctx, projectID)
	if err != nil {
		return err
	}
	s.client = client
	return nil
}

// SetFilter updates the log filter and refreshes
func (s *Service) SetFilter(filter string) {
	s.mu.Lock()
	s.filter = filter
	s.viewState = ViewList // Reset to list
	s.mu.Unlock()
}

// Update handles messages
func (s *Service) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "r":
			return s, s.Refresh()
		case "esc", "q":
			if s.viewState == ViewDetail {
				s.viewState = ViewList
				return s, nil
			}
			// Let parent handle exit if at root
		case "enter":
			if s.viewState == ViewList {
				if idx := s.table.Cursor(); idx >= 0 && idx < len(s.entries) {
					s.selectedLog = &s.entries[idx]
					s.viewState = ViewDetail
					// Initialize viewport content
					s.viewport = viewport.New(0, 0) // Dimensions set by WindowSizeMsg
					s.viewport.SetContent(s.renderLogDetails(s.selectedLog))
				}
			}
		}

	case tea.WindowSizeMsg:
		s.table.SetWidth(msg.Width - 4)
		s.table.SetHeight(msg.Height - 8)
		s.viewport.Width = msg.Width - 4
		s.viewport.Height = msg.Height - 8

	case []LogEntry:
		s.mu.Lock()
		s.entries = msg
		s.loading = false
		s.mu.Unlock()

		// Update Table
		rows := make([]table.Row, len(msg))
		for i, e := range msg {
			rows[i] = table.Row{
				e.Timestamp.Format("15:04:05"),
				e.Severity,
				e.ResourceID, // Show ID as it's more specific
				e.Payload,
			}
		}
		s.table.SetRows(rows)

	case error:
		s.err = msg
		s.loading = false
	}

	if s.viewState == ViewList {
		s.table, cmd = s.table.Update(msg)
	} else {
		s.viewport, cmd = s.viewport.Update(msg)
	}

	return s, cmd
}

// View renders the service UI
func (s *Service) View() string {
	if s.loading {
		return "Loading logs..."
	}
	if s.err != nil {
		return fmt.Sprintf("Error: %v", s.err)
	}

	if s.viewState == ViewList {
		header := styles.SubtleStyle.Render(fmt.Sprintf("Logging > Stream (Filter: %s)", s.filter))
		return lipgloss.JoinVertical(lipgloss.Left, header, s.table.View())
	}

	// Detail View
	if s.selectedLog != nil {
		header := styles.SubtleStyle.Render(fmt.Sprintf("Logging > Stream > %s", s.selectedLog.InsertID))
		return lipgloss.JoinVertical(lipgloss.Left, header, s.viewport.View())
	}

	return ""
}

// Refresh triggers data reload
func (s *Service) Refresh() tea.Cmd {
	s.loading = true
	return func() tea.Msg {
		if s.client == nil {
			return fmt.Errorf("logging client not initialized")
		}
		entries, err := s.client.ListEntries(s.ctx, s.filter)
		if err != nil {
			return err
		}
		return entries
	}
}

// HelpText returns keys
func (s *Service) HelpText() string {
	if s.viewState == ViewList {
		return "r:Refresh  Ent:Detail"
	}
	return "Esc/q:Back"
}

// Focus/Blur/Reset
func (s *Service) Focus() { s.table.Focus() }
func (s *Service) Blur()  { s.table.Blur() }
func (s *Service) Reset() {
	s.viewState = ViewList
	s.filter = "" // Or keep? Ideally reset if coming from fresh.
}
func (s *Service) IsRootView() bool { return s.viewState == ViewList }

func (s *Service) renderLogDetails(e *LogEntry) string {
	// Pretty print JSON if possible
	var prettyJSON string
	var obj interface{}
	if err := json.Unmarshal([]byte(e.FullPayload), &obj); err == nil {
		b, _ := json.MarshalIndent(obj, "", "  ")
		prettyJSON = string(b)
	} else {
		prettyJSON = e.FullPayload
	}

	return fmt.Sprintf(`
Timestamp: %s
Severity:  %s
Resource:  %s (%s)
InsertID:  %s

Payload:
%s
`,
		e.Timestamp,
		renderSeverity(e.Severity),
		e.Resource, e.ResourceID,
		e.InsertID,
		prettyJSON,
	)
}

func renderSeverity(sev string) string {
	switch sev {
	case "ERROR", "CRITICAL", "EMERGENCY", "ALERT":
		return styles.ErrorStyle.Render(sev)
	case "WARNING":
		return styles.WarningStyle.Render(sev)
	case "INFO":
		return styles.SuccessStyle.Render(sev)
	default:
		return styles.SubtleStyle.Render(sev)
	}
}
