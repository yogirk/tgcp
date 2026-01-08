package cloudrun

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

// -----------------------------------------------------------------------------
// Models
// -----------------------------------------------------------------------------

// Tick message for background refresh
type tickMsg time.Time

// ViewState defines the current UI state of the service
type ViewState int

type Tab int

const (
	TabServices Tab = iota
	TabFunctions
)

const (
	ViewList ViewState = iota
	ViewDetail
	ViewConfirmation
)

// servicesMsg is the message used to pass fetched data
type servicesMsg []RunService

// functionsMsg is the message used to pass fetched functions
type functionsMsg []Function

// errMsg is the standard error message
type errMsg error

// -----------------------------------------------------------------------------
// Service Definition
// -----------------------------------------------------------------------------

// Service implements the services.Service interface
type Service struct {
	client    *Client
	projectID string
	table     table.Model // Services Table
	funcTable table.Model // Functions Table

	// Tab Component
	activeTab Tab

	// UI Components
	filterInput textinput.Model
	filtering   bool

	// State
	services  []RunService
	functions []Function
	loading   bool
	err       error

	viewState       ViewState
	selectedService *RunService
	selectedFunc    *Function

	// Cache
	cache *core.Cache
}

// NewService creates a new instance of the service
func NewService(cache *core.Cache) *Service {
	// 1. Table Setup
	columns := []table.Column{
		{Title: "Name", Width: 30},
		{Title: "Status", Width: 10},
		{Title: "Region", Width: 15},
		{Title: "URL", Width: 50},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	// Apply styles
	s := table.DefaultStyles()
	s.Header = styles.HeaderStyle
	s.Selected = lipgloss.NewStyle().
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	// 1b. Functions Table Setup
	funcColumns := []table.Column{
		{Title: "Name", Width: 30},
		{Title: "Region", Width: 15},
		{Title: "State", Width: 10},
		{Title: "Updated", Width: 15},
	}
	ft := table.New(
		table.WithColumns(funcColumns),
		table.WithFocused(true),
		table.WithHeight(10),
	)
	ft.SetStyles(s)

	// 2. Filter Input Setup
	ti := textinput.New()
	ti.Placeholder = "Filter services..."
	ti.Prompt = "/ "
	ti.CharLimit = 100
	ti.Width = 50

	return &Service{
		table:       t,
		funcTable:   ft,
		activeTab:   TabServices,
		filterInput: ti,
		viewState:   ViewList,
		cache:       cache,
	}
}

// Name returns the full human-readable name
func (s *Service) Name() string {
	return "Cloud Run"
}

// ShortName returns the specialized identifier
func (s *Service) ShortName() string {
	return "run"
}

// HelpText returns context-aware keybindings
func (s *Service) HelpText() string {
	if s.viewState == ViewList {
		return "[]:Tabs  r:Refresh  /:Filter  Ent:Detail"
	}
	if s.viewState == ViewDetail {
		return "Esc/q:Back"
	}
	return ""
}

// -----------------------------------------------------------------------------
// Lifecycle & Interface Implementation
// -----------------------------------------------------------------------------

// InitService initializes the API client
func (s *Service) InitService(ctx context.Context, projectID string) error {
	s.projectID = projectID
	client, err := NewClient(ctx)
	if err != nil {
		return err
	}
	s.client = client
	return nil
}

// Init startup commands
func (s *Service) Init() tea.Cmd {
	return s.tick()
}

// tick creates a background ticker for cache invalidation
func (s *Service) tick() tea.Cmd {
	return tea.Tick(CacheTTL, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Refresh triggers a forced data reload
func (s *Service) Refresh() tea.Cmd {
	s.loading = true
	if s.activeTab == TabServices {
		return s.fetchDataCmd(true)
	} else {
		return s.fetchFunctionsCmd(true)
	}
}

// Reset clears the service state when navigating away or switching projects
func (s *Service) Reset() {
	s.viewState = ViewList
	s.selectedService = nil
	s.err = nil          // CRITICAL: Always clear errors on reset
	s.table.SetCursor(0) // Reset table position
	s.funcTable.SetCursor(0)
	s.activeTab = TabServices // Default to Services tab
}

// IsRootView returns true if we are at the top-level list
func (s *Service) IsRootView() bool {
	return s.viewState == ViewList
}

// Focus handles input focus (Visual Highlight)
func (s *Service) Focus() {
	s.table.Focus()
	st := table.DefaultStyles()
	st.Header = styles.HeaderStyle
	st.Selected = lipgloss.NewStyle().
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	s.table.SetStyles(st)
	s.funcTable.SetStyles(st)
}

// Blur handles loss of input focus (Visual Dimming)
func (s *Service) Blur() {
	s.table.Blur()
	st := table.DefaultStyles()
	st.Header = styles.HeaderStyle
	st.Selected = lipgloss.NewStyle().
		Foreground(styles.ColorText).
		Background(lipgloss.Color("237")). // Dark grey
		Bold(false)
	s.table.SetStyles(st)
	s.funcTable.SetStyles(st)
}

// -----------------------------------------------------------------------------
// Update Loop
// -----------------------------------------------------------------------------

func (s *Service) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	// 1. Background Tick
	case tickMsg:
		var batch []tea.Cmd
		if s.activeTab == TabServices {
			batch = append(batch, s.fetchDataCmd(false))
		} else {
			batch = append(batch, s.fetchFunctionsCmd(false))
		}
		batch = append(batch, s.tick())
		return s, tea.Batch(batch...)

	// 2. Data Loaded
	case servicesMsg:
		s.loading = false
		s.services = msg
		s.updateTable(s.services)
		return s, func() tea.Msg { return core.LastUpdatedMsg(time.Now()) }

	case functionsMsg:
		s.loading = false
		s.functions = msg
		s.updateFuncTable(s.functions)
		return s, func() tea.Msg { return core.LastUpdatedMsg(time.Now()) }

	// 3. Error Handling
	case errMsg:
		s.loading = false
		s.err = msg

	// 4. Window Resize
	case tea.WindowSizeMsg:
		const heightOffset = 6 // app header + status bar + padding
		newHeight := msg.Height - heightOffset
		if newHeight < 5 {
			newHeight = 5
		}
		s.table.SetHeight(newHeight)
		s.funcTable.SetHeight(newHeight)

	// 5. User Input
	case tea.KeyMsg:
		if s.viewState == ViewList {
			switch msg.String() {
			case "[", "]":
				// Switch Tab (Cycle)
				if s.activeTab == TabServices {
					s.activeTab = TabFunctions
					s.loading = true
					return s, s.fetchFunctionsCmd(true)
				} else {
					s.activeTab = TabServices
					return s, nil
				}
			case "r":
				return s, s.Refresh()
			case "enter":
				// Handle detail view selection
				if s.activeTab == TabServices {
					if s.selectedService == nil {
						// Check if table has selection
						svcs := s.services // Currently no filtering support in get logic, but if filtering added, use filtered
						if idx := s.table.Cursor(); idx >= 0 && idx < len(svcs) {
							s.selectedService = &svcs[idx]
							s.viewState = ViewDetail
						}
					}
				} else {
					// Function Details
					// Not fully implemented yet in prompt, but basic structure:
					// Just toggling detail view state is same.
					// Implementation plan said "Add function metadata viewer".
					// Let's implement basic select.
					funcs := s.functions
					if idx := s.funcTable.Cursor(); idx >= 0 && idx < len(funcs) {
						s.selectedFunc = &funcs[idx]
						s.viewState = ViewDetail
					}
				}
			}

			if s.activeTab == TabServices {
				s.table, cmd = s.table.Update(msg)
			} else {
				s.funcTable, cmd = s.funcTable.Update(msg)
			}
			return s, cmd
		} else if s.viewState == ViewDetail {
			switch msg.String() {
			case "q", "esc":
				s.viewState = ViewList
				s.selectedService = nil
				s.selectedFunc = nil
				return s, nil
			}
		}
	}

	return s, nil
}

// -----------------------------------------------------------------------------
// View Rendering
// -----------------------------------------------------------------------------

func (s *Service) View() string {
	if s.loading {
		return "Loading Cloud Run services..."
	}
	if s.err != nil {
		return fmt.Sprintf("Error fetching Cloud Run services: %v", s.err)
	}

	if s.viewState == ViewDetail {
		return s.renderDetailView()
	}

	// Default: List View
	return s.renderWithTabs()
}

func (s *Service) renderWithTabs() string {
	// Tabs
	var tabs string
	if s.activeTab == TabServices {
		tabs = lipgloss.JoinHorizontal(lipgloss.Top,
			styles.ActiveTabStyle.Render(" Services "),
			styles.InactiveTabStyle.Render(" Functions "),
		)
		return lipgloss.JoinVertical(lipgloss.Left, tabs, s.table.View())
	} else {
		tabs = lipgloss.JoinHorizontal(lipgloss.Top,
			styles.InactiveTabStyle.Render(" Services "),
			styles.ActiveTabStyle.Render(" Functions "),
		)
		return lipgloss.JoinVertical(lipgloss.Left, tabs, s.funcTable.View())
	}
}

func (s *Service) renderDetailView() string {
	if s.activeTab == TabFunctions {
		return s.renderFuncDetailView()
	}

	if s.selectedService == nil {
		return "No service selected"
	}

	svc := s.selectedService

	// Title
	// Title
	title := styles.SubtleStyle.Render(fmt.Sprintf("Cloud Run > Services > %s", svc.Name))

	// Status
	statusStyle := styles.SuccessStyle
	if svc.Status != StatusReady {
		statusStyle = styles.ErrorStyle
	}
	status := statusStyle.Render("● " + string(svc.Status))

	// Details
	details := fmt.Sprintf(`
%s %s
%s %s
%s %s
`,
		styles.LabelStyle.Render("Region:"), styles.ValueStyle.Render(svc.Region),
		styles.LabelStyle.Render("Status:"), status,
		styles.LabelStyle.Render("URL:"), styles.ValueStyle.Render(svc.URL),
	)

	// Wrap in a box
	content := lipgloss.JoinVertical(lipgloss.Left,
		title,
		details,
		"",
		styles.SubtleStyle.Render("Press 'q' or 'esc' to return"),
	)

	return styles.FocusedBoxStyle.Render(content)
}

func (s *Service) renderFuncDetailView() string {
	if s.selectedFunc == nil {
		return "No function selected"
	}
	f := s.selectedFunc
	title := styles.SubtleStyle.Render(fmt.Sprintf("Cloud Run > Functions > %s", f.Name))

	details := fmt.Sprintf(`
%s %s
%s %s
%s %s
`,
		styles.LabelStyle.Render("Region:"), styles.ValueStyle.Render(f.Region),
		styles.LabelStyle.Render("State:"), styles.ValueStyle.Render(f.State),
		styles.LabelStyle.Render("URL:"), styles.ValueStyle.Render(f.URL),
	)

	return styles.FocusedBoxStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
		title, details, "", styles.SubtleStyle.Render("Press 'q' or 'esc' to return"),
	))
}

// -----------------------------------------------------------------------------
// Helper Commands
// -----------------------------------------------------------------------------

func (s *Service) fetchDataCmd(force bool) tea.Cmd {
	return func() tea.Msg {
		key := "cloudrun_services"

		// 1. Check Cache
		if !force && s.cache != nil {
			if val, found := s.cache.Get(key); found {
				if svcs, ok := val.([]RunService); ok {
					return servicesMsg(svcs)
				}
			}
		}

		// 2. API Call
		if s.client == nil {
			return errMsg(fmt.Errorf("client not initialized"))
		}
		svcs, err := s.client.ListServices(s.projectID)
		if err != nil {
			return errMsg(err)
		}

		// 3. Update Cache
		if s.cache != nil {
			s.cache.Set(key, svcs, CacheTTL)
		}

		return servicesMsg(svcs)
	}
}

func (s *Service) updateTable(items []RunService) {
	rows := make([]table.Row, len(items))
	for i, item := range items {
		status := string(item.Status)
		if item.Status == StatusReady {
			status = "Ready"
		} else {
			status = string(item.Status)
		}
		rows[i] = table.Row{
			item.Name,
			status,
			item.Region,
			item.URL,
		}
	}
	s.table.SetRows(rows)
}

func (s *Service) fetchFunctionsCmd(force bool) tea.Cmd {
	return func() tea.Msg {
		key := "cloudrun_functions"
		if !force && s.cache != nil {
			if val, found := s.cache.Get(key); found {
				if items, ok := val.([]Function); ok {
					return functionsMsg(items)
				}
			}
		}
		if s.client == nil {
			return errMsg(fmt.Errorf("client not init"))
		}
		items, err := s.client.ListFunctions(s.projectID)
		if err != nil {
			return errMsg(err)
		}
		if s.cache != nil {
			s.cache.Set(key, items, CacheTTL)
		}
		return functionsMsg(items)
	}
}

func (s *Service) updateFuncTable(items []Function) {
	rows := make([]table.Row, len(items))
	for i, item := range items {
		state := item.State
		// Plain text state
		rows[i] = table.Row{
			item.Name,
			item.Region,
			state,
			item.LastUpdated.Format("2006-01-02"),
		}
	}
	s.funcTable.SetRows(rows)
}
