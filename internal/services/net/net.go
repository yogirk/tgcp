package net

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
	networksTable  table.Model
	subnetsTable   table.Model
	firewallsTable table.Model

	// UI
	filterInput textinput.Model
	activeTab   Tab

	// State
	networks  []Network
	subnets   []Subnet
	firewalls []Firewall
	loading   bool
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
	nTable := table.New(table.WithColumns(nCols), table.WithFocused(true), table.WithHeight(10))

	// Subnets Table
	sCols := []table.Column{
		{Title: "Name", Width: 30},
		{Title: "Region", Width: 15},
		{Title: "Range", Width: 15},
		{Title: "Gateway", Width: 15},
	}
	sTable := table.New(table.WithColumns(sCols), table.WithFocused(true), table.WithHeight(10))

	// Firewalls Table
	fCols := []table.Column{
		{Title: "Name", Width: 25},
		{Title: "Type", Width: 8},
		{Title: "Action", Width: 6},
		{Title: "Priority", Width: 8},
		{Title: "Source", Width: 20},
		{Title: "Target", Width: 20},
	}
	fTable := table.New(table.WithColumns(fCols), table.WithFocused(true), table.WithHeight(10))

	// Apply Styles
	s := table.DefaultStyles()
	s.Header = styles.HeaderStyle
	s.Selected = lipgloss.NewStyle().Foreground(lipgloss.Color("229")).Background(lipgloss.Color("57")).Bold(false)
	nTable.SetStyles(s)
	sTable.SetStyles(s)
	fTable.SetStyles(s)

	return &Service{
		networksTable:  nTable,
		subnetsTable:   sTable,
		firewallsTable: fTable,
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

func (s *Service) Init() tea.Cmd {
	return tea.Tick(CacheTTL, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func (s *Service) Refresh() tea.Cmd {
	s.loading = true
	if s.viewState == ViewList {
		return s.fetchNetworksCmd()
	} else if s.viewState == ViewDetail {
		// Refresh both for details
		return tea.Batch(s.fetchSubnetsCmd(), s.fetchFirewallsCmd())
	}
	return nil
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

	st := table.DefaultStyles()
	st.Header = styles.HeaderStyle
	st.Selected = lipgloss.NewStyle().Foreground(lipgloss.Color("229")).Background(lipgloss.Color("57")).Bold(false)
	s.networksTable.SetStyles(st)
	s.subnetsTable.SetStyles(st)
	s.firewallsTable.SetStyles(st)
}

func (s *Service) Blur() {
	s.networksTable.Blur()
	s.subnetsTable.Blur()
	s.firewallsTable.Blur()

	st := table.DefaultStyles()
	st.Header = styles.HeaderStyle
	st.Selected = lipgloss.NewStyle().Foreground(styles.ColorText).Background(lipgloss.Color("237")).Bold(false)
	s.networksTable.SetStyles(st)
	s.subnetsTable.SetStyles(st)
	s.firewallsTable.SetStyles(st)
}

// -----------------------------------------------------------------------------
// Update
// -----------------------------------------------------------------------------

func (s *Service) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tickMsg:
		if s.viewState == ViewList {
			return s, tea.Batch(s.fetchNetworksCmd(), s.Init())
		}
		return s, s.Init()

	case networksMsg:
		s.loading = false
		s.networks = msg
		s.updateNetworksTable()
		return s, func() tea.Msg { return core.LastUpdatedMsg(time.Now()) }

	case subnetsMsg:
		s.loading = false
		s.subnets = msg
		s.updateSubnetsTable()
		return s, func() tea.Msg { return core.LastUpdatedMsg(time.Now()) }

	case firewallsMsg:
		s.loading = false
		s.firewalls = msg
		s.updateFirewallsTable()
		return s, func() tea.Msg { return core.LastUpdatedMsg(time.Now()) }

	case errMsg:
		s.loading = false
		s.err = msg

	case tea.WindowSizeMsg:
		h := msg.Height - 6 // Header space
		if h < 5 {
			h = 5
		}
		s.networksTable.SetHeight(h)
		s.subnetsTable.SetHeight(h - 3) // Tab headers take extra space
		s.firewallsTable.SetHeight(h - 3)

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
					s.loading = true
					return s, tea.Batch(s.fetchSubnetsCmd(), s.fetchFirewallsCmd())
				}
			}
			s.networksTable, cmd = s.networksTable.Update(msg)
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

			if s.activeTab == TabSubnets {
				s.subnetsTable, cmd = s.subnetsTable.Update(msg)
				return s, cmd
			} else {
				s.firewallsTable, cmd = s.firewallsTable.Update(msg)
				return s, cmd
			}
		}
	}
	return s, nil
}

// -----------------------------------------------------------------------------
// View
// -----------------------------------------------------------------------------

func (s *Service) View() string {
	if s.loading {
		return "Loading Networking..."
	}
	if s.err != nil {
		return fmt.Sprintf("Error: %v", s.err)
	}

	if s.viewState == ViewList {
		return s.networksTable.View()
	} else if s.viewState == ViewDetail {
		return s.renderDetailView()
	}
	return ""
}

func (s *Service) renderDetailView() string {
	header := styles.SubtleStyle.Render(fmt.Sprintf("Networking > VPCs > %s", s.selectedNetwork.Name))

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
