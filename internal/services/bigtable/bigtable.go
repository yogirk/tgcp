package bigtable

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/yogirk/tgcp/internal/core"
	"github.com/yogirk/tgcp/internal/ui/components"
)

const CacheTTL = 60 * time.Second

// -----------------------------------------------------------------------------
// Models
// -----------------------------------------------------------------------------

type tickMsg time.Time

type ViewState int

const (
	ViewList ViewState = iota
	ViewDetail
)

type instancesMsg []Instance
type clustersMsg []Cluster
type errMsg error

// -----------------------------------------------------------------------------
// Service Definition
// -----------------------------------------------------------------------------

type Service struct {
	client    *Client
	projectID string
	table     *components.StandardTable

	filter components.FilterModel

	instances []Instance
	clusters  []Cluster // For selected instance
	loading   bool
	err       error

	viewState        ViewState
	selectedInstance *Instance

	cache *core.Cache
}

func NewService(cache *core.Cache) *Service {
	columns := []table.Column{
		{Title: "Name", Width: 30},
		{Title: "Display Name", Width: 30},
		{Title: "Type", Width: 15},
		{Title: "State", Width: 10},
	}

	t := components.NewStandardTable(columns)

	return &Service{
		table:     t,
		filter:     components.NewFilterWithPlaceholder("Filter instances..."),
		viewState: ViewList,
		cache:     cache,
	}
}

func (s *Service) Name() string {
	return "Bigtable"
}

func (s *Service) ShortName() string {
	return "bigtable"
}

func (s *Service) HelpText() string {
	if s.viewState == ViewList {
		return "r:Refresh  /:Filter  Ent:Detail"
	}
	return "Esc/q:Back"
}

// -----------------------------------------------------------------------------
// Lifecycle
// -----------------------------------------------------------------------------

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

func (s *Service) Refresh() tea.Cmd {
	s.loading = true
	return s.fetchInstancesCmd(true)
}

func (s *Service) Reset() {
	s.viewState = ViewList
	s.selectedInstance = nil
	s.clusters = nil
	s.err = nil
	s.table.SetCursor(0)
	s.filter.ExitFilterMode()
}

func (s *Service) IsRootView() bool {
	return s.viewState == ViewList
}

func (s *Service) Focus() {
	s.table.Focus()
}

func (s *Service) Blur() {
	s.table.Blur()
}

// -----------------------------------------------------------------------------
// Update
// -----------------------------------------------------------------------------

func (s *Service) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tickMsg:
		return s, tea.Batch(s.fetchInstancesCmd(false), s.tick())

	case instancesMsg:
		s.loading = false
		s.instances = msg
		s.updateTable(s.instances)
		return s, func() tea.Msg { return core.LastUpdatedMsg(time.Now()) }

	case clustersMsg:
		s.clusters = msg
		// We don't change state here, just store data for view

	case errMsg:
		s.loading = false
		s.err = msg

	case tea.WindowSizeMsg:
		s.table.HandleWindowSizeDefault(msg)

	case tea.KeyMsg:
		// Handle filter mode (only in list view)
		if s.viewState == ViewList {
			result := components.HandleFilterUpdate(
				&s.filter,
				msg,
				s.instances,
				func(items []Instance, query string) []Instance {
					return s.getFilteredInstances(items, query)
				},
				s.updateTable,
			)

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
			case "r":
				return s, s.Refresh()
			case "enter":
				instances := s.getFilteredInstances(s.instances, s.filter.Value())
				if idx := s.table.Cursor(); idx >= 0 && idx < len(instances) {
					s.selectedInstance = &instances[idx]
					s.viewState = ViewDetail
					s.clusters = nil // Clear old
					return s, s.fetchClustersCmd(s.selectedInstance.Name)
				}
			}
			var updatedTable *components.StandardTable
			updatedTable, cmd = s.table.Update(msg)
			s.table = updatedTable
			return s, cmd
		}

		if s.viewState == ViewDetail {
			switch msg.String() {
			case "esc", "q":
				s.viewState = ViewList
				s.selectedInstance = nil
				s.clusters = nil
				return s, nil
			}
		}
	}
	return s, nil
}

// -----------------------------------------------------------------------------
// Data & Helpers
// -----------------------------------------------------------------------------

func (s *Service) fetchInstancesCmd(force bool) tea.Cmd {
	return func() tea.Msg {
		key := fmt.Sprintf("bigtable:%s", s.projectID)
		if !force && s.cache != nil {
			if val, found := s.cache.Get(key); found {
				if items, ok := val.([]Instance); ok {
					return instancesMsg(items)
				}
			}
		}
		if s.client == nil {
			return errMsg(fmt.Errorf("client not init"))
		}
		items, err := s.client.ListInstances(s.projectID)
		if err != nil {
			return errMsg(err)
		}
		if s.cache != nil {
			s.cache.Set(key, items, CacheTTL)
		}
		return instancesMsg(items)
	}
}

func (s *Service) fetchClustersCmd(instanceID string) tea.Cmd {
	return func() tea.Msg {
		// No cache for clusters loop (simplification for MVP)
		if s.client == nil {
			return errMsg(fmt.Errorf("client not init"))
		}
		items, err := s.client.ListClusters(s.projectID, instanceID)
		if err != nil {
			return errMsg(err)
		}
		return clustersMsg(items)
	}
}

func (s *Service) updateTable(items []Instance) {
	rows := make([]table.Row, len(items))
	for i, item := range items {
		rows[i] = table.Row{
			item.Name,
			item.DisplayName,
			item.Type,
			item.State,
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
		return components.ContainsMatch(inst.Name, inst.DisplayName, inst.Type, inst.State)(q)
	})
}
