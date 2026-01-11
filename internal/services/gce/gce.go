package gce

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yogirk/tgcp/internal/core"
	"github.com/yogirk/tgcp/internal/ui/components"
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
	table     *components.StandardTable

	// UI Components
	filter        components.FilterModel
	filterSession components.FilterSession[Instance]
	spinner       components.SpinnerModel

	// State
	instances []Instance
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
	// Table Setup
	columns := GetGCEColumns()

	t := components.NewStandardTable(columns)

	svc := &Service{
		table:     t,
		filter:    components.NewFilterWithPlaceholder("Filter instances..."),
		spinner:   components.NewSpinner(),
		viewState: ViewList,
		cache:     cache,
	}
	svc.filterSession = components.NewFilterSession(&svc.filter, svc.getFilteredInstances, svc.updateTable)
	return svc
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

// Focus handles input focus
func (s *Service) Focus() {
	s.table.Focus()
}

// Blur handles loss of input focus
func (s *Service) Blur() {
	s.table.Blur()
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

// Reinit reinitializes the service with a new project ID
func (s *Service) Reinit(ctx context.Context, projectID string) error {
	s.Reset()
	return s.InitService(ctx, projectID)
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
	case components.SpinnerTickMsg:
		// Forward tick to spinner for animation
		s.spinner, cmd = s.spinner.Update(msg)
		return s, cmd

	case tickMsg:
		// Background refresh
		return s, tea.Batch(s.fetchInstancesCmd(false), s.tick())

	// Handle Data Fetching
	case instancesMsg:
		s.spinner.Stop()
		s.instances = msg
		s.filterSession.Apply(s.instances)
		return s, func() tea.Msg { return core.LastUpdatedMsg(time.Now()) }

	case errMsg:
		s.spinner.Stop()
		s.err = msg
		return s, nil

	case actionResultMsg:
		if msg.err != nil {
			s.err = msg.err
			// Show error toast
			return s, func() tea.Msg {
				return core.ToastMsg{Message: msg.err.Error(), Type: core.ToastError}
			}
		} else if msg.msg != "" {
			// Show success toast and refresh
			return s, tea.Batch(
				func() tea.Msg {
					return core.ToastMsg{Message: msg.msg, Type: core.ToastSuccess}
				},
				s.Refresh(),
			)
		} else {
			return s, s.Refresh()
		}

	case tea.WindowSizeMsg:
		s.table.HandleWindowSizeDefault(msg)

		// Optional: We could also resize columns here based on width
		// but let's stick to height for now to fix the "truncation" visual

	case tea.KeyMsg:
		// Handle filter mode (only in list view)
		if s.viewState == ViewList {
			result := s.filterSession.HandleKey(msg)

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

		// LIST VIEW KEYBINDINGS
		if s.viewState == ViewList {
			switch msg.String() {
			case "r":
				return s, s.fetchInstancesCmd(true)
			case "enter": // View Details
				instances := s.getFilteredInstances(s.instances, s.filter.Value())
				if idx := s.table.Cursor(); idx >= 0 && idx < len(instances) {
					s.selectedInstance = &instances[idx]
					s.viewState = ViewDetail
				}
			case "s": // Start (Confirm)
				instances := s.getFilteredInstances(s.instances, s.filter.Value())
				if idx := s.table.Cursor(); idx >= 0 && idx < len(instances) {
					s.selectedInstance = &instances[idx]
					s.pendingAction = "start"
					s.actionSource = ViewList
					s.viewState = ViewConfirmation
				}
			case "x": // Stop (Confirm)
				instances := s.getFilteredInstances(s.instances, s.filter.Value())
				if idx := s.table.Cursor(); idx >= 0 && idx < len(instances) {
					s.selectedInstance = &instances[idx]
					s.pendingAction = "stop"
					s.actionSource = ViewList
					s.viewState = ViewConfirmation
				}
			case "h": // SSH (Changed from Enter)
				instances := s.getFilteredInstances(s.instances, s.filter.Value())
				if idx := s.table.Cursor(); idx >= 0 && idx < len(instances) {
					return s, s.SSHCmd(instances[idx])
				}
			}
			// Forward to table
			var updatedTable *components.StandardTable
			updatedTable, cmd = s.table.Update(msg)
			s.table = updatedTable
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
	if s.err != nil {
		return components.RenderError(s.err, s.Name(), "Instances")
	}

	// Show animated spinner while loading
	if s.spinner.IsActive() {
		return s.spinner.View()
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
	return tea.Batch(
		s.spinner.Start(""), // Start animated spinner (empty = use playful messages)
		s.fetchInstancesCmd(false), // Smart refresh
	)
}

// Reset resets the service state
// Reset resets the service state
func (s *Service) Reset() {
	s.viewState = ViewList
	s.selectedInstance = nil
	s.err = nil          // Fix: Clear previous errors on reset
	s.table.SetCursor(0) // Optional: reset cursor to top
	s.filter.ExitFilterMode()
}

// IsRootView checks if we are in the main list view
func (s *Service) IsRootView() bool {
	return s.viewState == ViewList
}
