package net

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yogirk/tgcp/internal/core"
	"github.com/yogirk/tgcp/internal/styles"
	"github.com/yogirk/tgcp/internal/ui/components"
)

const CacheTTL = 5 * time.Minute

// -----------------------------------------------------------------------------
// Models
// -----------------------------------------------------------------------------

type tickMsg time.Time

type ViewState int

const (
	ViewList ViewState = iota
	ViewDetail
)

type Tab int

const (
	TabSubnets Tab = iota
	TabFirewalls
)

type networksMsg []Network
type subnetsMsg []Subnet
type firewallsMsg []Firewall
type errMsg error

// -----------------------------------------------------------------------------
// Service Definition
// -----------------------------------------------------------------------------

type Service struct {
	client    *Client
	projectID string

	// Tables
	networksTable  *components.StandardTable
	subnetsTable   *components.StandardTable
	firewallsTable *components.StandardTable

	// UI
	filterInput textinput.Model
	activeTab   Tab

	// State
	networks  []Network
	subnets   []Subnet
	firewalls []Firewall
	spinner   components.SpinnerModel
	err       error

	viewState       ViewState
	selectedNetwork *Network

	cache *core.Cache
}

func NewService(cache *core.Cache) *Service {
	// Networks Table
	nCols := []table.Column{
		{Title: "Name", Width: 30},
		{Title: "Mode", Width: 10},
		{Title: "IPv4 Range", Width: 20},
		{Title: "Gateway", Width: 15},
	}
	nTable := components.NewStandardTable(nCols)

	// Subnets Table
	sCols := []table.Column{
		{Title: "Name", Width: 30},
		{Title: "Region", Width: 15},
		{Title: "Range", Width: 15},
		{Title: "Gateway", Width: 15},
	}
	sTable := components.NewStandardTable(sCols)

	// Firewalls Table
	fCols := []table.Column{
		{Title: "Name", Width: 25},
		{Title: "Type", Width: 8},
		{Title: "Action", Width: 6},
		{Title: "Priority", Width: 8},
		{Title: "Source", Width: 20},
		{Title: "Target", Width: 20},
	}
	fTable := components.NewStandardTable(fCols)

	return &Service{
		networksTable:  nTable,
		subnetsTable:   sTable,
		firewallsTable: fTable,
		spinner:        components.NewSpinner(),
		viewState:      ViewList,
		activeTab:      TabSubnets,
		cache:          cache,
	}
}

func (s *Service) Name() string      { return "Networking" }
func (s *Service) ShortName() string { return "net" }

func (s *Service) HelpText() string {
	if s.viewState == ViewList {
		return "Ent:Detail  r:Refresh"
	}
	if s.viewState == ViewDetail {
		return "[]:Switch Tab  Esc/q:Back"
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
	return tea.Tick(CacheTTL, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func (s *Service) Refresh() tea.Cmd {
	var fetchCmd tea.Cmd
	if s.viewState == ViewList {
		fetchCmd = s.fetchNetworksCmd()
	} else if s.viewState == ViewDetail {
		fetchCmd = tea.Batch(s.fetchSubnetsCmd(), s.fetchFirewallsCmd())
	}
	if fetchCmd == nil {
		return nil
	}
	return tea.Batch(
		func() tea.Msg { return core.LoadingMsg{IsLoading: true} },
		fetchCmd,
		s.spinner.Start(""),
	)
}

func (s *Service) Reset() {
	s.viewState = ViewList
	s.selectedNetwork = nil
	s.activeTab = TabSubnets
	s.err = nil
	s.networksTable.SetCursor(0)
}

func (s *Service) IsRootView() bool {
	return s.viewState == ViewList
}

func (s *Service) Focus() {
	s.networksTable.Focus()
	s.subnetsTable.Focus()
	s.firewallsTable.Focus()
}

func (s *Service) Blur() {
	s.networksTable.Blur()
	s.subnetsTable.Blur()
	s.firewallsTable.Blur()
}

// -----------------------------------------------------------------------------
// Update
// -----------------------------------------------------------------------------

func (s *Service) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case components.SpinnerTickMsg:
		s.spinner, cmd = s.spinner.Update(msg)
		return s, cmd

	case tickMsg:
		if s.viewState == ViewList {
			return s, tea.Batch(s.fetchNetworksCmd(), s.Init())
		}
		return s, s.Init()

	case networksMsg:
		s.spinner.Stop()
		s.networks = msg
		s.updateNetworksTable()
		return s, tea.Batch(
			func() tea.Msg { return core.LoadingMsg{IsLoading: false} },
			func() tea.Msg { return core.LastUpdatedMsg(time.Now()) },
		)

	case subnetsMsg:
		s.spinner.Stop()
		s.subnets = msg
		s.updateSubnetsTable()
		return s, tea.Batch(
			func() tea.Msg { return core.LoadingMsg{IsLoading: false} },
			func() tea.Msg { return core.LastUpdatedMsg(time.Now()) },
		)

	case firewallsMsg:
		s.spinner.Stop()
		s.firewalls = msg
		s.updateFirewallsTable()
		return s, tea.Batch(
			func() tea.Msg { return core.LoadingMsg{IsLoading: false} },
			func() tea.Msg { return core.LastUpdatedMsg(time.Now()) },
		)

	case errMsg:
		s.spinner.Stop()
		s.err = msg
		return s, func() tea.Msg { return core.LoadingMsg{IsLoading: false} }

	case tea.WindowSizeMsg:
		s.networksTable.HandleWindowSizeDefault(msg)
		// Tab headers take extra space for detail view tables
		s.subnetsTable.HandleWindowSize(msg, 9)
		s.firewallsTable.HandleWindowSize(msg, 9)

	case tea.KeyMsg:
		switch msg.String() {
		case "r":
			return s, s.Refresh()
		}

		if s.viewState == ViewList {
			if msg.String() == "enter" {
				if s.networksTable.Cursor() >= 0 && s.networksTable.Cursor() < len(s.networks) {
					s.selectedNetwork = &s.networks[s.networksTable.Cursor()]
					s.viewState = ViewDetail
					s.activeTab = TabSubnets // Default to subnets
					return s, tea.Batch(s.fetchSubnetsCmd(), s.fetchFirewallsCmd(), s.spinner.Start(""))
				}
			}
			var updatedTable *components.StandardTable
			updatedTable, cmd = s.networksTable.Update(msg)
			s.networksTable = updatedTable
			return s, cmd

		} else if s.viewState == ViewDetail {
			switch msg.String() {
			case "esc", "q":
				s.viewState = ViewList
				s.selectedNetwork = nil
				return s, nil
			case "[", "]", "tab": // Allow tab-like switching
				if s.activeTab == TabSubnets {
					s.activeTab = TabFirewalls
				} else {
					s.activeTab = TabSubnets
				}
				return s, nil
			}

			var updatedTable *components.StandardTable
			if s.activeTab == TabSubnets {
				updatedTable, cmd = s.subnetsTable.Update(msg)
				s.subnetsTable = updatedTable
			} else {
				updatedTable, cmd = s.firewallsTable.Update(msg)
				s.firewallsTable = updatedTable
			}
			return s, cmd
		}
	}
	return s, nil
}

// -----------------------------------------------------------------------------
// View
// -----------------------------------------------------------------------------

func (s *Service) View() string {
	if s.err != nil {
		return components.RenderError(s.err, s.Name(), "Networks")
	}

	// Show spinner while loading
	if s.spinner.IsActive() {
		return s.spinner.View()
	}

	if s.viewState == ViewList {
		breadcrumb := components.Breadcrumb(
			fmt.Sprintf("Project %s", s.projectID),
			s.Name(),
			"Networks",
		)
		return lipgloss.JoinVertical(lipgloss.Left, breadcrumb, s.networksTable.View())
	} else if s.viewState == ViewDetail {
		return s.renderDetailView()
	}
	return ""
}

func (s *Service) renderDetailView() string {
	header := components.Breadcrumb(
		fmt.Sprintf("Project %s", s.projectID),
		s.Name(),
		"Networks",
		s.selectedNetwork.Name,
	)

	// Tabs
	var subStyle, fwStyle lipgloss.Style
	if s.activeTab == TabSubnets {
		subStyle = styles.ActiveTabStyle
		fwStyle = styles.InactiveTabStyle
	} else {
		subStyle = styles.InactiveTabStyle
		fwStyle = styles.ActiveTabStyle
	}

	tabs := lipgloss.JoinHorizontal(lipgloss.Top,
		subStyle.Render("Subnets"),
		fwStyle.Render("Firewall Rules"),
	)

	var content string
	if s.activeTab == TabSubnets {
		content = s.subnetsTable.View()
	} else {
		content = s.firewallsTable.View()
	}

	return lipgloss.JoinVertical(lipgloss.Left, header, tabs, content)
}

// -----------------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------------

func (s *Service) fetchNetworksCmd() tea.Cmd {
	return func() tea.Msg {
		if s.client == nil {
			return errMsg(fmt.Errorf("client not init"))
		}
		nets, err := s.client.ListNetworks(s.projectID)
		if err != nil {
			return errMsg(err)
		}
		return networksMsg(nets)
	}
}

func (s *Service) fetchSubnetsCmd() tea.Cmd {
	return func() tea.Msg {
		if s.selectedNetwork == nil {
			return errMsg(fmt.Errorf("no network"))
		}
		subs, err := s.client.ListSubnets(s.projectID, s.selectedNetwork.SelfLink)
		if err != nil {
			return errMsg(err)
		}
		return subnetsMsg(subs)
	}
}

func (s *Service) fetchFirewallsCmd() tea.Cmd {
	return func() tea.Msg {
		if s.selectedNetwork == nil {
			return errMsg(fmt.Errorf("no network"))
		}
		fws, err := s.client.ListFirewalls(s.projectID, s.selectedNetwork.SelfLink)
		if err != nil {
			return errMsg(err)
		}
		return firewallsMsg(fws)
	}
}

func (s *Service) updateNetworksTable() {
	rows := make([]table.Row, len(s.networks))
	for i, n := range s.networks {
		mode := n.Mode
		if mode == "AUTO" {
			mode = "AUTO"
		}
		rows[i] = table.Row{n.Name, mode, n.IPv4Range, n.GatewayIPv4}
	}
	s.networksTable.SetRows(rows)
}

func (s *Service) updateSubnetsTable() {
	rows := make([]table.Row, len(s.subnets))
	for i, sub := range s.subnets {
		rows[i] = table.Row{sub.Name, sub.Region, sub.IPCidrRange, sub.Gateway}
	}
	s.subnetsTable.SetRows(rows)
}

func (s *Service) updateFirewallsTable() {
	rows := make([]table.Row, len(s.firewalls))
	for i, f := range s.firewalls {
		prio := fmt.Sprintf("%d", f.Priority)
		action := f.Action
		if action == "ALLOW" {
			action = "ALLOW"
		} else if action == "DENY" {
			action = "DENY"
		}
		rows[i] = table.Row{f.Name, f.Direction, action, prio, f.Source, f.Target}
	}
	s.firewallsTable.SetRows(rows)
}
