package dataproc

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
const DefaultRegion = "us-central1" // Simplification for MVP

// -----------------------------------------------------------------------------
// Models
// -----------------------------------------------------------------------------

type tickMsg time.Time

type ViewState int

const (
	ViewList ViewState = iota
	ViewDetail
)

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

	clusters []Cluster
	loading  bool
	err      error

	viewState       ViewState
	selectedCluster *Cluster

	cache *core.Cache
}

func NewService(cache *core.Cache) *Service {
	columns := []table.Column{
		{Title: "Cluster Name", Width: 30},
		{Title: "Status", Width: 15},
		{Title: "Workers", Width: 10},
		{Title: "Zone", Width: 15},
	}

	t := components.NewStandardTable(columns)

	return &Service{
		table:     t,
		filter:     components.NewFilterWithPlaceholder("Filter clusters..."),
		viewState: ViewList,
		cache:     cache,
	}
}

func (s *Service) Name() string {
	return "Dataproc"
}

func (s *Service) ShortName() string {
	return "dataproc"
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
	return s.fetchClustersCmd(true)
}

func (s *Service) Reset() {
	s.viewState = ViewList
	s.selectedCluster = nil
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
		return s, tea.Batch(s.fetchClustersCmd(false), s.tick())

	case clustersMsg:
		s.loading = false
		s.clusters = msg
		s.updateTable(s.clusters)
		return s, func() tea.Msg { return core.LastUpdatedMsg(time.Now()) }

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
				s.clusters,
				func(items []Cluster, query string) []Cluster {
					return s.getFilteredClusters(items, query)
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
				clusters := s.getFilteredClusters(s.clusters, s.filter.Value())
				if idx := s.table.Cursor(); idx >= 0 && idx < len(clusters) {
					s.selectedCluster = &clusters[idx]
					s.viewState = ViewDetail
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
				s.selectedCluster = nil
				return s, nil
			}
		}
	}
	return s, nil
}

// -----------------------------------------------------------------------------
// Data & Helpers
// -----------------------------------------------------------------------------

func (s *Service) fetchClustersCmd(force bool) tea.Cmd {
	return func() tea.Msg {
		key := fmt.Sprintf("dataproc:%s:%s", s.projectID, DefaultRegion)
		if !force && s.cache != nil {
			if val, found := s.cache.Get(key); found {
				if items, ok := val.([]Cluster); ok {
					return clustersMsg(items)
				}
			}
		}
		if s.client == nil {
			return errMsg(fmt.Errorf("client not init"))
		}
		items, err := s.client.ListClusters(s.projectID, DefaultRegion)
		if err != nil {
			return errMsg(err)
		}
		if s.cache != nil {
			s.cache.Set(key, items, CacheTTL)
		}
		return clustersMsg(items)
	}
}

func (s *Service) updateTable(items []Cluster) {
	rows := make([]table.Row, len(items))
	for i, item := range items {
		rows[i] = table.Row{
			item.Name,
			item.Status,
			fmt.Sprintf("%d", item.WorkerCount),
			item.Zone,
		}
	}
	s.table.SetRows(rows)
}

// getFilteredClusters returns filtered clusters based on the query string
func (s *Service) getFilteredClusters(clusters []Cluster, query string) []Cluster {
	if query == "" {
		return clusters
	}
	return components.FilterSlice(clusters, query, func(cluster Cluster, q string) bool {
		return components.ContainsMatch(cluster.Name, cluster.Status, cluster.Zone)(q)
	})
}
