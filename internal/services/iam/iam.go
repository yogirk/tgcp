package iam

import (
	"context"

	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rk/tgcp/internal/core"
	"github.com/rk/tgcp/internal/styles"
)

const CacheTTL = 5 * time.Minute

// Tick message for background refresh
type tickMsg time.Time

// Service implements the generic Service interface for IAM
type Service struct {
	client    *Client
	projectID string
	table     table.Model

	// State
	accounts []ServiceAccount
	loading  bool
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

	return &Service{
		table: t,
		cache: cache,
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

// Msg types
type accountsMsg []ServiceAccount
type errMsg error

func (s *Service) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tickMsg:
		return s, tea.Batch(s.fetchAccountsCmd(false), s.tick())

	case accountsMsg:
		s.loading = false
		s.accounts = msg
		s.updateTable(msg)
		return s, func() tea.Msg { return core.LastUpdatedMsg(time.Now()) }

	case errMsg:
		s.loading = false
		s.err = msg

	case tea.WindowSizeMsg:
		const heightOffset = 6
		newHeight := msg.Height - heightOffset
		if newHeight < 5 {
			newHeight = 5
		}
		s.table.SetHeight(newHeight)

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
			s.table, cmd = s.table.Update(msg)
			return s, cmd
		}
	}

	return s, nil
}

func (s *Service) View() string {
	if s.loading {
		return "Loading Service Accounts..."
	}
	if s.err != nil {
		return "Error: " + s.err.Error()
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
	s.loading = true
	return s.fetchAccountsCmd(false)
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
