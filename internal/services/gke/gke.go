package gke

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/yogirk/tgcp/internal/core"
	"github.com/yogirk/tgcp/internal/ui/components"
)

const CacheTTL = 60 * time.Second

// -----------------------------------------------------------------------------
// Models & Msgs
// -----------------------------------------------------------------------------

// Tick message for background refresh
type tickMsg time.Time

// ViewState defines the current UI state of the service
type ViewState int

const (
	ViewList ViewState = iota
	ViewDetail
	ViewConfirmation
)

// Msg types
type clustersMsg []Cluster
type errMsg error
type actionResultMsg struct {
	err error
	msg string
}

// -----------------------------------------------------------------------------
// Service Definition
// -----------------------------------------------------------------------------

type Service struct {
	client    *Client
	projectID string
	table     *components.StandardTable

	// UI Components
	filter        components.FilterModel
	filterSession components.FilterSession[Cluster]
	spinner       components.SpinnerModel

	// State
	clusters []Cluster
	err      error

	// View State
	viewState       ViewState
	selectedCluster *Cluster

	// Confirmation State
	pendingAction string    // e.g. "connect"
	actionSource  ViewState // Where to return after confirmation

	// Cache
	cache *core.Cache
}

func NewService(cache *core.Cache) *Service {
	// 1. Table Setup
	columns := []table.Column{
		{Title: "Name", Width: 25},
		{Title: "Location", Width: 15},
		{Title: "Status", Width: 12},
		{Title: "Version", Width: 18},
		{Title: "Mode", Width: 10},
		{Title: "Nodes", Width: 8},
	}

	t := components.NewStandardTable(columns)

	svc := &Service{
		table:     t,
		filter:    components.NewFilterWithPlaceholder("Filter clusters..."),
		spinner:   components.NewSpinner(),
		viewState: ViewList,
		cache:     cache,
	}
	svc.filterSession = components.NewFilterSession(&svc.filter, svc.getFilteredClusters, svc.updateTable)
	return svc
}

func (s *Service) Name() string {
	return "Kubernetes Engine"
}

func (s *Service) ShortName() string {
	return "gke"
}

func (s *Service) HelpText() string {
	if s.viewState == ViewList {
		return "r:Refresh  /:Filter  K:k9s  l:Logs  Ent:Detail"
	}
	if s.viewState == ViewDetail {
		return "Esc/q:Back  K:k9s"
	}
	if s.viewState == ViewConfirmation {
		return "y:Confirm  n:Cancel"
	}
	return ""
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
	return tea.Batch(
		s.spinner.Start(""),
		s.fetchClustersCmd(true),
	)
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
// Update Loop
// -----------------------------------------------------------------------------

func (s *Service) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case components.SpinnerTickMsg:
		s.spinner, cmd = s.spinner.Update(msg)
		return s, cmd

	case tickMsg:
		return s, tea.Batch(s.fetchClustersCmd(false), s.tick())

	case clustersMsg:
		s.spinner.Stop()
		s.clusters = msg
		s.filterSession.Apply(s.clusters)
		return s, func() tea.Msg { return core.LastUpdatedMsg(time.Now()) }

	case errMsg:
		s.spinner.Stop()
		s.err = msg
		return s, nil

	case actionResultMsg:
		if msg.err != nil {
			s.err = msg.err
			return s, func() tea.Msg {
				return core.ToastMsg{Message: msg.err.Error(), Type: core.ToastError}
			}
		} else if msg.msg != "" {
			return s, func() tea.Msg {
				return core.ToastMsg{Message: msg.msg, Type: core.ToastSuccess}
			}
		}
		return s, nil

	case tea.WindowSizeMsg:
		s.table.HandleWindowSizeDefault(msg)

	case tea.MouseMsg:
		// Forward mouse events to table for click selection
		if s.viewState == ViewList {
			var updatedTable *components.StandardTable
			updatedTable, cmd = s.table.Update(msg)
			s.table = updatedTable
			return s, cmd
		}

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

		// LIST VIEW
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
			case "K": // Launch k9s
				clusters := s.getFilteredClusters(s.clusters, s.filter.Value())
				if idx := s.table.Cursor(); idx >= 0 && idx < len(clusters) {
					c := clusters[idx]
					return s, s.launchK9s(c)
				}
			case "l": // Logs
				clusters := s.getFilteredClusters(s.clusters, s.filter.Value())
				if idx := s.table.Cursor(); idx >= 0 && idx < len(clusters) {
					c := clusters[idx]
					// Filter for GKE Cluster logs
					filter := fmt.Sprintf(`resource.type="k8s_cluster" AND resource.labels.cluster_name="%s" AND resource.labels.location="%s"`, c.Name, c.Location)
					heading := fmt.Sprintf("Cluster: %s", c.Name)
					return s, func() tea.Msg { return core.SwitchToLogsMsg{Filter: filter, Source: "gke", Heading: heading} }
				}
			}
			var updatedTable *components.StandardTable
			updatedTable, cmd = s.table.Update(msg)
			s.table = updatedTable
			return s, cmd
		}

		// DETAIL VIEW
		if s.viewState == ViewDetail {
			switch msg.String() {
			case "esc", "q":
				s.viewState = ViewList
				s.selectedCluster = nil
				return s, nil
			case "K": // Launch k9s
				if s.selectedCluster != nil {
					return s, s.launchK9s(*s.selectedCluster)
				}
			}
		}

		// CONFIRMATION VIEW
		// (Not currently used but ready)
		if s.viewState == ViewConfirmation {
			switch msg.String() {
			case "y", "enter":
				// s.pendingAction logic
				s.viewState = s.actionSource
				s.pendingAction = ""
				return s, nil
			case "n", "esc", "q":
				s.viewState = s.actionSource
				s.pendingAction = ""
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
	if s.err != nil {
		return components.RenderError(s.err, s.Name(), "Clusters")
	}

	// Show animated spinner while loading
	if s.spinner.IsActive() {
		return s.spinner.View()
	}

	if s.viewState == ViewDetail {
		return s.renderDetailView()
	}

	return s.renderListView()
}

func (s *Service) renderListView() string {
	// Filter Bar
	var content strings.Builder
	content.WriteString(components.Breadcrumb(
		fmt.Sprintf("Project %s", s.projectID),
		s.Name(),
		"Clusters",
	))
	content.WriteString("\n")
	content.WriteString(s.filter.View())
	content.WriteString("\n")
	content.WriteString(s.table.View())
	return content.String()
}

// -----------------------------------------------------------------------------
// Helper Commands
// -----------------------------------------------------------------------------

func (s *Service) fetchClustersCmd(force bool) tea.Cmd {
	return func() tea.Msg {
		key := fmt.Sprintf("gke_clusters:%s", s.projectID)

		if !force && s.cache != nil {
			if val, found := s.cache.Get(key); found {
				if items, ok := val.([]Cluster); ok {
					return clustersMsg(items)
				}
			}
		}

		if s.client == nil {
			return errMsg(fmt.Errorf("client not initialized"))
		}

		clusters, err := s.client.ListClusters(s.projectID)
		if err != nil {
			return errMsg(err)
		}

		if s.cache != nil {
			s.cache.Set(key, clusters, CacheTTL)
		}

		return clustersMsg(clusters)
	}
}

func (s *Service) updateTable(items []Cluster) {
	rows := make([]table.Row, len(items))
	for i, item := range items {
		status := item.Status
		if item.Status == "RUNNING" {
			status = "RUNNING" // could add color here but table handles it poorly
		}

		rows[i] = table.Row{
			item.Name,
			item.Location,
			status,
			item.MasterVersion,
			item.Mode,
			fmt.Sprintf("%d", item.NodeCount),
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
		return components.ContainsMatch(cluster.Name, cluster.Location, cluster.Status, cluster.MasterVersion)(q)
	})
}

func (s *Service) launchK9s(c Cluster) tea.Cmd {
	return tea.ExecProcess(exec.Command("k9s", "--context", fmt.Sprintf("gke_%s_%s_%s", s.projectID, c.Location, c.Name)), func(err error) tea.Msg {
		if err != nil {
			// Fallback: Try to get credentials first?
			// The user might not have context set up.
			// Best effort: Run gcloud get-credentials then k9s
			cmdStr := fmt.Sprintf("gcloud container clusters get-credentials %s --zone %s --project %s && k9s", c.Name, c.Location, s.projectID)
			return tea.ExecProcess(exec.Command("bash", "-c", cmdStr), func(err error) tea.Msg {
				if err != nil {
					return actionResultMsg{err: err}
				}
				return actionResultMsg{msg: "k9s session ended"}
			})
		}
		return actionResultMsg{msg: "k9s session ended"}
	})
}
