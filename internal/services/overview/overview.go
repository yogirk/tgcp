package overview

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yogirk/tgcp/internal/core"
)

const CacheTTL = 15 * time.Minute // Longer TTL for billing/inventory data

// -----------------------------------------------------------------------------
// Service Definition
// -----------------------------------------------------------------------------

type Service struct {
	client    *Client
	projectID string
	data      DashboardData
	cache     *core.Cache
}

func NewService(cache *core.Cache) *Service {
	return &Service{
		cache: cache,
		data: DashboardData{
			InfoLoading:      true,
			RecsLoading:      true,
			InventoryLoading: true,
			// BudgetsLoading: true, // dependent on Info
		},
	}
}

func (s *Service) Name() string {
	return "Overview"
}

func (s *Service) ShortName() string {
	return "overview"
}

func (s *Service) HelpText() string {
	return "r:Refresh"
}

func (s *Service) Refresh() tea.Cmd {
	s.data.InfoLoading = true
	s.data.RecsLoading = true
	s.data.InventoryLoading = true
	s.data.BudgetsLoading = true

	// Manual refresh forces cache clear
	s.cache.Delete(fmt.Sprintf("billing:recs:%s", s.projectID))
	s.cache.Delete(fmt.Sprintf("billing:inventory:global:%s", s.projectID))
	if s.data.Info.BillingAccountID != "" {
		s.cache.Delete(fmt.Sprintf("billing:budgets:%s", s.data.Info.BillingAccountID))
	}

	cmds := []tea.Cmd{
		s.fetchInfoCmd(),
		s.fetchRecsCmd(),
		s.fetchInventoryCmd(),
	}

	if s.data.Info.BillingAccountID != "" {
		cmds = append(cmds, s.fetchBudgetsCmd(s.data.Info.BillingAccountID))
	}

	return tea.Batch(cmds...)
}

// isLoading returns true if any data is still loading
func (s *Service) isLoading() bool {
	return s.data.InfoLoading || s.data.RecsLoading || s.data.InventoryLoading || s.data.BudgetsLoading
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
	return s.Refresh()
}

func (s *Service) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case InfoMsg:
		s.data.Info = BillingInfo(msg)
		s.data.InfoLoading = false
		// Once we have info (and account ID), we can load budgets
		if s.data.Info.BillingAccountID != "" {
			return s, s.fetchBudgetsCmd(s.data.Info.BillingAccountID)
		}
		s.data.BudgetsLoading = false

	case RecsMsg:
		s.data.Recommendations = []Recommendation(msg)
		s.data.RecsLoading = false

	case InventoryMsg:
		s.data.Inventory = ResourceInventory(msg)
		s.data.InventoryLoading = false

	case BudgetsMsg:
		s.data.Budgets = []SpendLimit(msg)
		s.data.BudgetsLoading = false

	case ErrMsg:
		s.data.Error = msg
		// Clear loaders on critical error
		s.data.InfoLoading = false
		s.data.RecsLoading = false
		s.data.InventoryLoading = false
		s.data.BudgetsLoading = false

	case tea.KeyMsg:
		switch msg.String() {
		case "r":
			return s, s.Refresh()
		}
	}
	return s, nil
}

// View implementation is in views.go

// -----------------------------------------------------------------------------
// Messages
// -----------------------------------------------------------------------------

type ErrMsg error

// Helpers

func (s *Service) fetchInfoCmd() tea.Cmd {
	return func() tea.Msg {
		if s.client == nil {
			return ErrMsg(fmt.Errorf("client not init"))
		}
		info, err := s.client.GetProjectBillingInfo(s.projectID)
		if err != nil {
			return ErrMsg(err)
		}
		return InfoMsg(info)
	}
}

func (s *Service) fetchRecsCmd() tea.Cmd {
	return func() tea.Msg {
		if s.client == nil {
			return RecsMsg{}
		}

		// Cache Check
		cacheKey := fmt.Sprintf("billing:recs:%s", s.projectID)
		if val, found := s.cache.Get(cacheKey); found {
			if recs, ok := val.([]Recommendation); ok {
				return RecsMsg(recs)
			}
		}

		recs, err := s.client.GetRecommendations(s.projectID, "")
		if err != nil {
			return RecsMsg{}
		}

		// Cache Set
		s.cache.Set(cacheKey, recs, CacheTTL)

		return RecsMsg(recs)
	}
}

func (s *Service) fetchInventoryCmd() tea.Cmd {
	return func() tea.Msg {
		if s.client == nil {
			return InventoryMsg{}
		}

		// Cache Check
		cacheKey := fmt.Sprintf("billing:inventory:global:%s", s.projectID)
		if val, found := s.cache.Get(cacheKey); found {
			if inv, ok := val.(ResourceInventory); ok {
				return InventoryMsg(inv)
			}
		}

		inv, err := s.client.GetGlobalInventory(s.projectID)
		if err != nil {
			return InventoryMsg{}
		}

		// Cache Set
		s.cache.Set(cacheKey, inv, CacheTTL)

		return InventoryMsg(inv)
	}
}

func (s *Service) fetchBudgetsCmd(billingAccountID string) tea.Cmd {
	return func() tea.Msg {
		if s.client == nil || billingAccountID == "" {
			return BudgetsMsg{}
		}

		cacheKey := fmt.Sprintf("billing:budgets:%s", billingAccountID)
		if val, found := s.cache.Get(cacheKey); found {
			if b, ok := val.([]SpendLimit); ok {
				return BudgetsMsg(b)
			}
		}

		budgets, err := s.client.GetBudgets(billingAccountID)
		if err != nil {
			return BudgetsMsg{}
		}

		s.cache.Set(cacheKey, budgets, CacheTTL)

		return BudgetsMsg(budgets)
	}
}

// Reset clears state
func (s *Service) Reset() {
}

// Focus/Blur stubs
func (s *Service) Focus()           {}
func (s *Service) Blur()            {}
func (s *Service) IsRootView() bool { return true }
