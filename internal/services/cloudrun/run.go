package cloudrun

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yogirk/tgcp/internal/core"
	"github.com/yogirk/tgcp/internal/styles"
	"github.com/yogirk/tgcp/internal/ui/components"
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
	table     *components.StandardTable // Services Table
	funcTable *components.StandardTable // Functions Table

	// Tab Component
	activeTab Tab

	// UI Components
	filter                components.FilterModel
	serviceFilterSession  components.FilterSession[RunService]
	functionFilterSession components.FilterSession[Function]

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

	t := components.NewStandardTable(columns)

	// 1b. Functions Table Setup
	funcColumns := []table.Column{
		{Title: "Name", Width: 30},
		{Title: "Region", Width: 15},
		{Title: "State", Width: 10},
		{Title: "Updated", Width: 15},
	}
	ft := components.NewStandardTable(funcColumns)

	svc := &Service{
		table:     t,
		funcTable: ft,
		activeTab: TabServices,
		filter:    components.NewFilterWithPlaceholder("Filter services..."),
		viewState: ViewList,
		cache:     cache,
	}
	svc.serviceFilterSession = components.NewFilterSession(&svc.filter, svc.getFilteredServices, svc.updateTable)
	svc.functionFilterSession = components.NewFilterSession(&svc.filter, svc.getFilteredFunctions, svc.updateFuncTable)
	return svc
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

// Reinit reinitializes the service with a new project ID
func (s *Service) Reinit(ctx context.Context, projectID string) error {
	s.Reset()
	return s.InitService(ctx, projectID)
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
	s.filter.ExitFilterMode()
}

// IsRootView returns true if we are at the top-level list
func (s *Service) IsRootView() bool {
	return s.viewState == ViewList
}

// Focus handles input focus (Visual Highlight)
func (s *Service) Focus() {
	s.table.Focus()
	s.funcTable.Focus()
}

// Blur handles loss of input focus (Visual Dimming)
func (s *Service) Blur() {
	s.table.Blur()
	s.funcTable.Blur()
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
		s.serviceFilterSession.Apply(s.services)
		return s, func() tea.Msg { return core.LastUpdatedMsg(time.Now()) }

	case functionsMsg:
		s.loading = false
		s.functions = msg
		s.functionFilterSession.Apply(s.functions)
		return s, func() tea.Msg { return core.LastUpdatedMsg(time.Now()) }

	// 3. Error Handling
	case errMsg:
		s.loading = false
		s.err = msg

	// 4. Window Resize
	case tea.WindowSizeMsg:
		s.table.HandleWindowSizeDefault(msg)
		s.funcTable.HandleWindowSizeDefault(msg)

	// 5. User Input
	case tea.KeyMsg:
		// Handle filter mode (only in list view)
		if s.viewState == ViewList {
			var result components.FilterUpdateResult
			if s.activeTab == TabServices {
				result = s.serviceFilterSession.HandleKey(msg)
			} else {
				result = s.functionFilterSession.HandleKey(msg)
			}

			if result.Handled {
				if result.Cmd != nil {
					return s, result.Cmd
				}
				if !result.ShouldContinue {
					return s, nil
				}
				// Continue processing other keys
			}
		}

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
					s.serviceFilterSession.Apply(s.services)
					return s, nil
				}
			case "r":
				return s, s.Refresh()
			case "/":
				// Enter filter mode
				cmd := s.filter.EnterFilterMode()
				return s, cmd
			case "enter":
				// Handle detail view selection
				if s.activeTab == TabServices {
					if s.selectedService == nil {
						svcs := s.getFilteredServices(s.services, s.filter.Value())
						if idx := s.table.Cursor(); idx >= 0 && idx < len(svcs) {
							s.selectedService = &svcs[idx]
							s.viewState = ViewDetail
						}
					}
				} else {
					funcs := s.getFilteredFunctions(s.functions, s.filter.Value())
					if idx := s.funcTable.Cursor(); idx >= 0 && idx < len(funcs) {
						s.selectedFunc = &funcs[idx]
						s.viewState = ViewDetail
					}
				}
			}

			var updatedTable *components.StandardTable
			if s.activeTab == TabServices {
				updatedTable, cmd = s.table.Update(msg)
				s.table = updatedTable
			} else {
				updatedTable, cmd = s.funcTable.Update(msg)
				s.funcTable = updatedTable
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
		return components.RenderSpinner("Loading Cloud Run services...")
	}
	if s.err != nil {
		return components.RenderError(s.err, s.Name(), "Services")
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
	var tableView string

	// Filter Bar
	var filterBar string
	filterBar = s.filter.View() + "\n"

	if s.activeTab == TabServices {
		tabs = lipgloss.JoinHorizontal(lipgloss.Top,
			styles.ActiveTabStyle.Render(" Services "),
			styles.InactiveTabStyle.Render(" Functions "),
		)
		tableView = s.table.View()
	} else {
		tabs = lipgloss.JoinHorizontal(lipgloss.Top,
			styles.InactiveTabStyle.Render(" Services "),
			styles.ActiveTabStyle.Render(" Functions "),
		)
		tableView = s.funcTable.View()
	}

	return lipgloss.JoinVertical(lipgloss.Left, tabs, filterBar, tableView)
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

// getFilteredServices returns filtered services based on the query string
func (s *Service) getFilteredServices(services []RunService, query string) []RunService {
	if query == "" {
		return services
	}
	return components.FilterSlice(services, query, func(svc RunService, q string) bool {
		return components.ContainsMatch(svc.Name, string(svc.Status), svc.Region, svc.URL)(q)
	})
}

// getFilteredFunctions returns filtered functions based on the query string
func (s *Service) getFilteredFunctions(functions []Function, query string) []Function {
	if query == "" {
		return functions
	}
	return components.FilterSlice(functions, query, func(fn Function, q string) bool {
		return components.ContainsMatch(fn.Name, fn.Region, fn.State, fn.URL)(q)
	})
}
