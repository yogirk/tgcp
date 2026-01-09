package gke

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rk/tgcp/internal/core"
	"github.com/rk/tgcp/internal/styles"
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
type actionResultMsg struct{ err error }

// -----------------------------------------------------------------------------
// Service Definition
// -----------------------------------------------------------------------------

type Service struct {
	client    *Client
	projectID string
	table     table.Model

	// UI Components
	filterInput textinput.Model
	filtering   bool

	// State
	clusters []Cluster
	loading  bool
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

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()
	s.Header = styles.HeaderStyle
	s.Selected = lipgloss.NewStyle().
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	// 2. Filter Input Setup
	ti := textinput.New()
	ti.Placeholder = "Filter clusters..."
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
	return "Kubernetes Engine"
}

func (s *Service) ShortName() string {
	return "gke"
}

func (s *Service) HelpText() string {
	if s.viewState == ViewList {
		return "r:Refresh  /:Filter  K:k9s  Ent:Detail"
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
	return s.fetchClustersCmd(true) // force refresh
}

func (s *Service) Reset() {
	s.viewState = ViewList
	s.selectedCluster = nil
	s.err = nil
	s.table.SetCursor(0)
	s.filtering = false
	s.filterInput.Reset()
}

func (s *Service) IsRootView() bool {
	return s.viewState == ViewList
}

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

func (s *Service) Blur() {
	s.table.Blur()
	st := table.DefaultStyles()
	st.Header = styles.HeaderStyle
	st.Selected = lipgloss.NewStyle().
		Foreground(styles.ColorText).
		Background(lipgloss.Color("237")).
		Bold(false)
	s.table.SetStyles(st)
}

// -----------------------------------------------------------------------------
// Update Loop
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

	case actionResultMsg:
		if msg.err != nil {
			s.err = msg.err
		}
		// Don't auto verify/refresh for external commands usually, but ok.
		return s, nil

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
				s.filtering = false
				s.filterInput.Blur()
				s.filterInput.Reset()
				s.updateTable(s.clusters)
				return s, nil
			case "enter":
				s.filtering = false
				s.filterInput.Blur()
				return s, nil
			}
			var inputCmd tea.Cmd
			s.filterInput, inputCmd = s.filterInput.Update(msg)
			// TODO: Implement actual string matching filter if needed
			return s, inputCmd
		}

		// LIST VIEW
		if s.viewState == ViewList {
			switch msg.String() {
			case "/":
				s.filtering = true
				s.filterInput.Focus()
				return s, textinput.Blink
			case "r":
				return s, s.Refresh()
			case "enter":
				if idx := s.table.Cursor(); idx >= 0 && idx < len(s.clusters) {
					s.selectedCluster = &s.clusters[idx]
					s.viewState = ViewDetail
				}
			case "K": // Launch k9s
				if idx := s.table.Cursor(); idx >= 0 && idx < len(s.clusters) {
					c := s.clusters[idx]
					return s, s.launchK9s(c)
				}
			}
			s.table, cmd = s.table.Update(msg)
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
	if s.loading && len(s.clusters) == 0 {
		return "Loading Kubernetes clusters..."
	}
	if s.err != nil {
		return fmt.Sprintf("Error: %v", s.err)
	}

	if s.viewState == ViewDetail {
		return s.renderDetailView()
	}

	return s.renderListView()
}

func (s *Service) renderListView() string {
	return s.table.View()
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

func (s *Service) launchK9s(c Cluster) tea.Cmd {
	return tea.ExecProcess(exec.Command("k9s", "--context", fmt.Sprintf("gke_%s_%s_%s", s.projectID, c.Location, c.Name)), func(err error) tea.Msg {
		if err != nil {
			// Fallback: Try to get credentials first?
			// The user might not have context set up.
			// Best effort: Run gcloud get-credentials then k9s
			cmdStr := fmt.Sprintf("gcloud container clusters get-credentials %s --zone %s --project %s && k9s", c.Name, c.Location, s.projectID)
			return tea.ExecProcess(exec.Command("bash", "-c", cmdStr), func(err error) tea.Msg {
				return actionResultMsg{err}
			})
		}
		return actionResultMsg{nil}
	})
}
