package iam

import (
	"context"

	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/yogirk/tgcp/internal/core"
	"github.com/yogirk/tgcp/internal/ui/components"
)

const CacheTTL = 5 * time.Minute

// Tick message for background refresh
type tickMsg time.Time

// Service implements the generic Service interface for IAM
type Service struct {
	client    *Client
	projectID string
	table     *components.StandardTable

	// State
	accounts []ServiceAccount
	spinner  components.SpinnerModel
	err      error

	// View State
	viewDetail      bool
	selectedAccount *ServiceAccount

	// Cache
	cache *core.Cache
}

func NewService(cache *core.Cache) *Service {
	// Table Setup
	columns := []table.Column{
		{Title: "Display Name", Width: 30},
		{Title: "Email", Width: 40},
		{Title: "Status", Width: 10},
		{Title: "ID", Width: 25},
	}

	t := components.NewStandardTable(columns)

	return &Service{
		table:   t,
		spinner: components.NewSpinner(),
		cache:   cache,
	}
}

func (s *Service) Name() string {
	return "Identity & Access Management"
}

func (s *Service) ShortName() string {
	return "iam"
}

func (s *Service) HelpText() string {
	if s.viewDetail {
		return "Esc/q:Back"
	}
	return "r:Refresh  Ent:Detail"
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
type accountsMsg []ServiceAccount
type errMsg error

func (s *Service) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case components.SpinnerTickMsg:
		s.spinner, cmd = s.spinner.Update(msg)
		return s, cmd

	case tickMsg:
		return s, tea.Batch(s.fetchAccountsCmd(false), s.tick())

	case accountsMsg:
		s.spinner.Stop()
		s.accounts = msg
		s.updateTable(msg)
		return s, func() tea.Msg { return core.LastUpdatedMsg(time.Now()) }

	case errMsg:
		s.spinner.Stop()
		s.err = msg
		return s, nil

	case tea.WindowSizeMsg:
		s.table.HandleWindowSizeDefault(msg)

	case tea.MouseMsg:
		// Forward mouse events to table for click selection
		if !s.viewDetail {
			var updatedTable *components.StandardTable
			updatedTable, cmd = s.table.Update(msg)
			s.table = updatedTable
			return s, cmd
		}

	case tea.KeyMsg:
		if s.viewDetail {
			// Detail View Keybindings
			switch msg.String() {
			case "esc", "q":
				s.viewDetail = false
				s.selectedAccount = nil
				return s, nil
			}
		} else {
			// List View Keybindings
			switch msg.String() {
			case "r":
				return s, s.fetchAccountsCmd(true)
			case "enter":
				if idx := s.table.Cursor(); idx >= 0 && idx < len(s.accounts) {
					s.selectedAccount = &s.accounts[idx]
					s.viewDetail = true
				}
			}
			var updatedTable *components.StandardTable
			updatedTable, cmd = s.table.Update(msg)
			s.table = updatedTable
			return s, cmd
		}
	}

	return s, nil
}

func (s *Service) View() string {
	if s.err != nil {
		return components.RenderError(s.err, s.Name(), "Service Accounts")
	}

	// Show spinner while loading
	if s.spinner.IsActive() {
		return s.spinner.View()
	}

	if s.viewDetail {
		return s.renderDetailView()
	}

	return s.renderServiceAccountsList()
}

// Cmd to fetch accounts
func (s *Service) fetchAccountsCmd(force bool) tea.Cmd {
	return func() tea.Msg {
		key := "iam_accounts"

		// 1. Check Cache
		if !force && s.cache != nil {
			if val, found := s.cache.Get(key); found {
				if accs, ok := val.([]ServiceAccount); ok {
					return accountsMsg(accs)
				}
			}
		}

		// 2. API Call
		if s.client == nil {
			return errMsg(fmt.Errorf("client not initialized"))
		}
		accs, err := s.client.ListServiceAccounts(s.projectID)
		if err != nil {
			return errMsg(err)
		}

		// 3. Update Cache
		if s.cache != nil {
			s.cache.Set(key, accs, CacheTTL)
		}

		return accountsMsg(accs)
	}
}

func (s *Service) Refresh() tea.Cmd {
	return tea.Batch(
		s.spinner.Start(""),
		s.fetchAccountsCmd(false),
	)
}

func (s *Service) Reset() {
	s.viewDetail = false
	s.selectedAccount = nil
	s.err = nil // Fix: Clear previous errors on reset
	s.table.SetCursor(0)
}

func (s *Service) IsRootView() bool {
	return !s.viewDetail
}

// Internal Helpers

func (s *Service) updateTable(accounts []ServiceAccount) {
	rows := make([]table.Row, len(accounts))
	for i, acc := range accounts {
		status := "Active"
		if acc.Disabled {
			status = "Disabled"
		}

		rows[i] = table.Row{
			acc.DisplayName,
			acc.Email,
			status,
			acc.UniqueID,
		}
	}
	s.table.SetRows(rows)
}
