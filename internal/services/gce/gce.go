package gce

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

const CacheTTL = 30 * time.Second

// Tick message for background refresh
type tickMsg time.Time

// ViewState defines whether we are listing or viewing details
type ViewState int

const (
	ViewList ViewState = iota
	ViewDetail
	ViewConfirmation
)

// Service implements the services.Service interface for GCE
type Service struct {
	client    *Client
	projectID string
	table     table.Model

	// UI Components
	filterInput textinput.Model
	filtering   bool

	// State
	instances []Instance
	filtered  []Instance
	loading   bool
	err       error

	// View State
	viewState        ViewState
	selectedInstance *Instance

	// Confirmation State
	pendingAction string    // "start" or "stop"
	actionSource  ViewState // Where to return after confirmation

	// Cache
	cache *core.Cache
}

func NewService(cache *core.Cache) *Service {
	// Table Setup (same as before)
	// Table Setup
	columns := []table.Column{
		{Title: "Name", Width: 30},
		{Title: "Status", Width: 10},
		{Title: "Zone", Width: 12},
		{Title: "Internal IP", Width: 15},
		{Title: "External IP", Width: 15},
		{Title: "ID", Width: 20},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()
	s.Header = styles.HeaderStyle
	s.Selected = styles.SelectedItemStyle.Copy().UnsetBorderLeft() // Reuse existing style but tweak it?
	// Or just use a simple style for table selection
	s.Selected = lipgloss.NewStyle().
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)

	t.SetStyles(s)

	// Filter Input Setup
	ti := textinput.New()
	ti.Placeholder = "Filter instances..."
	ti.Prompt = "/ "
	ti.CharLimit = 100
	ti.Width = 50

	return &Service{
		table:       t,
		filterInput: ti,
		viewState:   ViewList,
		cache:       cache,
	}
}

func (s *Service) Name() string {
	return "Google Compute Engine"
}

func (s *Service) ShortName() string {
	return "gce"
}

func (s *Service) HelpText() string {
	if s.viewState == ViewList {
		return "r:Refresh  /:Filter  s:Start  x:Stop  h:SSH  Ent:Detail"
	}
	if s.viewState == ViewDetail {
		return "Esc/q:Back  s:Start  x:Stop  h:SSH"
	}
	if s.viewState == ViewConfirmation {
		return "y:Confirm  n:Cancel"
	}
	return ""
}

// Msg types
type instancesMsg []Instance
type errMsg error

// InitService initializes the service logic (API clients)
func (s *Service) InitService(ctx context.Context, projectID string) error {
	s.projectID = projectID
	client, err := NewClient(ctx)
	if err != nil {
		return err
	}
	s.client = client
	return nil
}

// Init satisfies tea.Model interface
func (s *Service) Init() tea.Cmd {
	return s.tick()
}

func (s *Service) tick() tea.Cmd {
	return tea.Tick(CacheTTL, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Update handles messages specific to GCE
func (s *Service) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tickMsg:
		// Background refresh
		// Only refresh if we are visible? Or always?
		// For now, allow refresh.
		// Use false to allow cache hit if multiple ticks stack up,
		// but actually we want to force check validity or re-fetch?
		// Actually tick means "TTL might be up".
		// But our fetches set the Cache with TTL.
		// If we call fetch(false), it will check cache. If cache expired, it fetches.
		return s, tea.Batch(s.fetchInstancesCmd(false), s.tick())

	// Handle Data Fetching
	case instancesMsg:
		s.loading = false
		s.instances = msg
		s.updateTable(s.instances)
		return s, func() tea.Msg { return core.LastUpdatedMsg(time.Now()) }

	case errMsg:
		s.loading = false
		s.err = msg

	case actionResultMsg:
		if msg.err != nil {
			s.err = msg.err
		} else {
			// Trigger refresh after action or just show status?
			// For MVP, simple re-fetch after a delay might be nice,
			// but let's just refresh now.
			return s, s.Refresh()
		}

	case tea.WindowSizeMsg:
		// Calculate available height for the table
		// Height - StatusBar(1) - Padding(2) - TableHeader(1) - ExtraGap(2)
		const heightOffset = 6
		// Width - Sidebar(25) - Padding(4)
		const widthOffset = 30

		newHeight := msg.Height - heightOffset
		if newHeight < 5 {
			newHeight = 5 // Minimum height
		}

		s.table.SetHeight(newHeight)

		// Optional: We could also resize columns here based on width
		// but let's stick to height for now to fix the "truncation" visual

	case tea.KeyMsg:
		// FILTERING MODE
		if s.filtering {
			switch msg.String() {
			case "esc":
				s.filtering = false
				s.filterInput.Blur()
				s.filterInput.Reset()
				s.updateTable(s.instances) // Reset table
				return s, nil
			case "enter":
				s.filtering = false
				s.filterInput.Blur()
				// Keep filtered results
				return s, nil
			}

			var inputCmd tea.Cmd
			s.filterInput, inputCmd = s.filterInput.Update(msg)

			// Update Filter Logic
			query := s.filterInput.Value()
			s.filterInstances(query)

			return s, inputCmd
		}

		// LIST VIEW KEYBINDINGS
		if s.viewState == ViewList {
			switch msg.String() {
			case "/":
				s.filtering = true
				s.filterInput.Focus()
				return s, textinput.Blink
			case "r":
				return s, s.fetchInstancesCmd(true)
			case "enter": // View Details
				instances := s.getCurrentInstances()
				if idx := s.table.Cursor(); idx >= 0 && idx < len(instances) {
					s.selectedInstance = &instances[idx]
					s.viewState = ViewDetail
				}
			case "s": // Start (Confirm)
				instances := s.getCurrentInstances()
				if idx := s.table.Cursor(); idx >= 0 && idx < len(instances) {
					s.selectedInstance = &instances[idx]
					s.pendingAction = "start"
					s.actionSource = ViewList
					s.viewState = ViewConfirmation
				}
			case "x": // Stop (Confirm)
				instances := s.getCurrentInstances()
				if idx := s.table.Cursor(); idx >= 0 && idx < len(instances) {
					s.selectedInstance = &instances[idx]
					s.pendingAction = "stop"
					s.actionSource = ViewList
					s.viewState = ViewConfirmation
				}
			case "h": // SSH (Changed from Enter)
				instances := s.getCurrentInstances()
				if idx := s.table.Cursor(); idx >= 0 && idx < len(instances) {
					return s, s.SSHCmd(instances[idx])
				}
			}
			// Forward to table
			s.table, cmd = s.table.Update(msg)
			return s, cmd
		}

		// DETAIL VIEW KEYBINDINGS
		if s.viewState == ViewDetail {
			switch msg.String() {
			case "esc", "q":
				// Return to List View
				s.viewState = ViewList
				s.selectedInstance = nil
				return s, nil
			case "s": // Start (Confirm)
				if s.selectedInstance != nil {
					s.pendingAction = "start"
					s.actionSource = ViewDetail
					s.viewState = ViewConfirmation
				}
			case "x": // Stop (Confirm)
				if s.selectedInstance != nil {
					s.pendingAction = "stop"
					s.actionSource = ViewDetail
					s.viewState = ViewConfirmation
				}
			case "h": // SSH
				if s.selectedInstance != nil {
					return s, s.SSHCmd(*s.selectedInstance)
				}
			}
			// No other updates needed for static detail view
		}

		// CONFIRMATION VIEW KEYBINDINGS
		if s.viewState == ViewConfirmation {
			switch msg.String() {
			case "y", "enter": // Confirm
				var actionCmd tea.Cmd
				if s.pendingAction == "start" {
					actionCmd = s.StartInstanceCmd(*s.selectedInstance)
				} else if s.pendingAction == "stop" {
					actionCmd = s.StopInstanceCmd(*s.selectedInstance)
				}

				// Return to source view
				s.viewState = s.actionSource
				// We might want to keep selectedInstance if we return to ViewDetail
				// If we return to ViewList, selectedInstance is usually kept too until explicitly cleared or changed

				// Reset pending state
				s.pendingAction = ""
				return s, actionCmd

			case "n", "esc", "q": // Cancel
				s.viewState = s.actionSource
				s.pendingAction = ""
				return s, nil
			}
		}

	// Handle Table Events (only in List View)
	default:
		if s.viewState == ViewList {
			s.table, cmd = s.table.Update(msg)
		}
	}

	return s, cmd
}

// View renders the service UI
func (s *Service) View() string {
	if s.loading {
		return "Loading instances..."
	}
	if s.err != nil {
		return fmt.Sprintf("Error: %v", s.err)
	}

	if s.viewState == ViewDetail {
		return s.renderDetailView()
	}

	if s.viewState == ViewConfirmation {
		return s.renderConfirmation()
	}

	// Default: List View
	return s.renderListView()
}

// Cmd to fetch instances
func (s *Service) fetchInstancesCmd(force bool) tea.Cmd {
	return func() tea.Msg {
		key := "gce_instances"

		// 1. Check Cache
		if !force && s.cache != nil {
			if val, found := s.cache.Get(key); found {
				if insts, ok := val.([]Instance); ok {
					return instancesMsg(insts)
				}
			}
		}

		// 2. API Call
		if s.client == nil {
			// This can happen if Refresh is called before InitService
			return errMsg(fmt.Errorf("client not initialized"))
		}
		insts, err := s.client.ListInstances(s.projectID)
		if err != nil {
			return errMsg(err)
		}

		// 3. Update Cache
		if s.cache != nil {
			s.cache.Set(key, insts, CacheTTL)
		}

		return instancesMsg(insts)
	}
}

// Cmd to refresh (public)
func (s *Service) Refresh() tea.Cmd {
	s.loading = true
	return s.fetchInstancesCmd(false) // Smart refresh
}

// Reset resets the service state
func (s *Service) Reset() {
	s.viewState = ViewList
	s.selectedInstance = nil
	s.table.SetCursor(0) // Optional: reset cursor to top
}

// IsRootView checks if we are in the main list view
func (s *Service) IsRootView() bool {
	return s.viewState == ViewList
}
