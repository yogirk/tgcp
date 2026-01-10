package cloudsql

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

const CacheTTL = 60 * time.Second

// Tick message for background refresh
type tickMsg time.Time

// ViewState defines whether we are listing or viewing details
type ViewState int

const (
	ViewList ViewState = iota
	ViewDetail
	ViewConfirmation
)

// Service implements the generic Service interface
type Service struct {
	client    *Client
	projectID string
	table     *components.StandardTable

	// UI Components
	filter        components.FilterModel
	filterSession components.FilterSession[Instance]

	// State
	instances []Instance
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
	// Table Setup
	columns := []table.Column{
		{Title: "Name", Width: 50},
		{Title: "Status", Width: 15},
		{Title: "Version", Width: 20},
		{Title: "Region", Width: 15},
		{Title: "Primary IP", Width: 20},
		{Title: "Tier", Width: 20},
	}

	t := components.NewStandardTable(columns)

	svc := &Service{
		table:     t,
		filter:    components.NewFilterWithPlaceholder("Filter instances..."),
		viewState: ViewList,
		cache:     cache,
	}
	svc.filterSession = components.NewFilterSession(&svc.filter, svc.getFilteredInstances, svc.updateTable)
	return svc
}

func (s *Service) Name() string {
	return "Cloud SQL"
}

func (s *Service) ShortName() string {
	return "sql"
}

func (s *Service) HelpText() string {
	if s.viewState == ViewList {
		return "r:Refresh  /:Filter  s:Start  x:Stop  Ent:Detail"
	}
	if s.viewState == ViewDetail {
		return "Esc/q:Back  s:Start  x:Stop"
	}
	if s.viewState == ViewConfirmation {
		return "y:Confirm  n:Cancel"
	}
	return ""
}

func (s *Service) Focus() {
	s.table.Focus()
}

func (s *Service) Blur() {
	s.table.Blur()
}

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

func (s *Service) Init() tea.Cmd {
	return s.tick()
}

func (s *Service) tick() tea.Cmd {
	return tea.Tick(CacheTTL, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Msg types
type instancesMsg []Instance
type errMsg error
type actionResultMsg struct{ err error }

func (s *Service) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tickMsg:
		return s, tea.Batch(s.fetchInstancesCmd(false), s.tick())

	case instancesMsg:
		s.loading = false
		s.instances = msg
		s.filterSession.Apply(s.instances)
		return s, func() tea.Msg { return core.LastUpdatedMsg(time.Now()) }

	case errMsg:
		s.loading = false
		s.err = msg

	case actionResultMsg:
		if msg.err != nil {
			s.err = msg.err
		}
		// Refresh after action
		return s, s.Refresh()

	case tea.WindowSizeMsg:
		s.table.HandleWindowSizeDefault(msg)

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

		switch s.viewState {
		case ViewList:
			switch msg.String() {
			case "r":
				return s, s.fetchInstancesCmd(true)
			case "enter":
				instances := s.getFilteredInstances(s.instances, s.filter.Value())
				if idx := s.table.Cursor(); idx >= 0 && idx < len(instances) {
					s.selectedInstance = &instances[idx]
					s.viewState = ViewDetail
				}
			case "s": // Start
				instances := s.getFilteredInstances(s.instances, s.filter.Value())
				if idx := s.table.Cursor(); idx >= 0 && idx < len(instances) {
					s.selectedInstance = &instances[idx]
					s.pendingAction = "start"
					s.actionSource = ViewList
					s.viewState = ViewConfirmation
				}
			case "x": // Stop
				instances := s.getFilteredInstances(s.instances, s.filter.Value())
				if idx := s.table.Cursor(); idx >= 0 && idx < len(instances) {
					s.selectedInstance = &instances[idx]
					s.pendingAction = "stop"
					s.actionSource = ViewList
					s.viewState = ViewConfirmation
				}
			}
			var updatedTable *components.StandardTable
			updatedTable, cmd = s.table.Update(msg)
			s.table = updatedTable
			return s, cmd

		case ViewDetail:
			switch msg.String() {
			case "esc", "q":
				s.viewState = ViewList
				s.selectedInstance = nil
				return s, nil
			case "s":
				if s.selectedInstance != nil {
					s.pendingAction = "start"
					s.actionSource = ViewDetail
					s.viewState = ViewConfirmation
				}
			case "x":
				if s.selectedInstance != nil {
					s.pendingAction = "stop"
					s.actionSource = ViewDetail
					s.viewState = ViewConfirmation
				}
			}

		case ViewConfirmation:
			switch msg.String() {
			case "y", "enter":
				var actionCmd tea.Cmd
				if s.pendingAction == "start" {
					actionCmd = s.startInstanceCmd(*s.selectedInstance)
				} else if s.pendingAction == "stop" {
					actionCmd = s.stopInstanceCmd(*s.selectedInstance)
				}
				s.viewState = s.actionSource
				s.pendingAction = ""
				return s, actionCmd

			case "n", "esc", "q":
				s.viewState = s.actionSource
				s.pendingAction = ""
				return s, nil
			}
		}
	}

	return s, nil
}

func (s *Service) View() string {
	if s.loading {
		return components.RenderSpinner("Loading Cloud SQL instances...")
	}
	if s.err != nil {
		return components.RenderError(s.err, s.Name(), "Instances")
	}

	if s.viewState == ViewDetail {
		return s.renderDetailView()
	}
	if s.viewState == ViewConfirmation {
		return s.renderConfirmation()
	}

	// Filter Bar
	var content strings.Builder
	content.WriteString(components.Breadcrumb(
		fmt.Sprintf("Project %s", s.projectID),
		s.Name(),
		"Instances",
	))
	content.WriteString("\n")
	content.WriteString(s.filter.View())
	content.WriteString("\n")
	content.WriteString(s.table.View())
	return content.String()
}

// Cmd to fetch instances
func (s *Service) fetchInstancesCmd(force bool) tea.Cmd {
	return func() tea.Msg {
		key := "sql_instances"

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

func (s *Service) Refresh() tea.Cmd {
	s.loading = true
	return s.fetchInstancesCmd(false)
}

func (s *Service) Reset() {
	s.viewState = ViewList
	s.selectedInstance = nil
	s.err = nil // Fix: Clear previous errors on reset
	s.table.SetCursor(0)
	s.filter.ExitFilterMode()
}

func (s *Service) IsRootView() bool {
	return s.viewState == ViewList
}

// Internal Helpers

func (s *Service) updateTable(instances []Instance) {
	rows := make([]table.Row, len(instances))
	for i, inst := range instances {
		state := string(inst.State)
		if state == "" {
			state = "UNKNOWN"
		}

		if state == "RUNNABLE" {
			state = "RUNNABLE"
		} else if state == "STOPPED" || state == "FAILED" {
			state = string(inst.State)
		} else {
			state = string(inst.State)
		}

		rows[i] = table.Row{
			inst.Name,
			state,
			inst.DatabaseVersion,
			inst.Region,
			inst.PrimaryIP,
			inst.Tier,
		}
	}
	s.table.SetRows(rows)
}

// getFilteredInstances returns filtered instances based on the query string
func (s *Service) getFilteredInstances(instances []Instance, query string) []Instance {
	if query == "" {
		return instances
	}
	return components.FilterSlice(instances, query, func(inst Instance, q string) bool {
		return components.ContainsMatch(inst.Name, string(inst.State), inst.DatabaseVersion, inst.Region, inst.PrimaryIP, inst.Tier)(q)
	})
}

func (s *Service) startInstanceCmd(i Instance) tea.Cmd {
	return func() tea.Msg {
		err := s.client.StartInstance(s.projectID, i.Name)
		return actionResultMsg{err}
	}
}

func (s *Service) stopInstanceCmd(i Instance) tea.Cmd {
	return func() tea.Msg {
		err := s.client.StopInstance(s.projectID, i.Name)
		return actionResultMsg{err}
	}
}
