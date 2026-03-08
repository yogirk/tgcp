package cloudbuild

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yogirk/tgcp/internal/core"
	"github.com/yogirk/tgcp/internal/ui/components"
)

const CacheTTL = 30 * time.Second

// =============================================================================
// Models
// =============================================================================

// BuildItem represents a resource managed by this service
type BuildItem struct {
	Name   string
	Status string
	ID     string
	Region string
}

// Tick message for background refresh
type tickMsg time.Time

// ViewState defines the current UI state of the service
type ViewState int

const (
	ViewList ViewState = iota
	ViewDetail
	ViewConfirmation
)

// Message types for async operations
type dataMsg []BuildItem
type errMsg error
type actionResultMsg struct {
	action string // "start", "stop", etc.
	name   string // resource name for toast
	err    error
}

// =============================================================================
// Service Definition
// =============================================================================

// Service implements the services.Service interface
type Service struct {
	client    *Client
	projectID string

	// Dimensions (for custom layouts)
	width  int
	height int

	// UI Components
	table         *components.StandardTable
	filter        components.FilterModel
	filterSession components.FilterSession[BuildItem]
	spinner       components.SpinnerModel

	// Data State
	items   []BuildItem
	err     error
	loaded  bool // Track if initial data has been loaded

	// View State
	viewState    ViewState
	selectedItem *BuildItem

	// Confirmation State
	pendingAction string
	actionSource  ViewState

	// Cache
	cache *core.Cache
}

// Placeholder Client type - replace with actual GCP client
type Client struct{}

func NewClient(ctx context.Context) (*Client, error) {
	// Initialize your GCP client here
	return &Client{}, nil
}

// NewService creates a new instance of the service
func NewService(cache *core.Cache) *Service {
	columns := []table.Column{
		{Title: "Name", Width: 30},
		{Title: "Status", Width: 15},
		{Title: "Region", Width: 15},
		{Title: "ID", Width: 20},
	}

	t := components.NewStandardTable(columns)

	svc := &Service{
		table:     t,
		filter:    components.NewFilterWithPlaceholder("Filter items..."),
		spinner:   components.NewSpinner(),
		viewState: ViewList,
		cache:     cache,
		loaded:    false,
	}
	svc.filterSession = components.NewFilterSession(&svc.filter, svc.getFilteredItems, svc.updateTable)
	return svc
}

// Name returns the full human-readable name
func (s *Service) Name() string {
	return "Cloud Build"
}

// ShortName returns the identifier used for routing (e.g., "gce", "sql")
func (s *Service) ShortName() string {
	return "cloudbuild"
}

// HelpText returns context-aware keybindings for the status bar
func (s *Service) HelpText() string {
	switch s.viewState {
	case ViewList:
		return "r:Refresh  /:Filter  Enter:Detail"
	case ViewDetail:
		return "Esc/q:Back"
	case ViewConfirmation:
		return "y:Confirm  n:Cancel"
	default:
		return ""
	}
}

// =============================================================================
// Lifecycle & Interface Implementation
// =============================================================================

// InitService initializes the API client - called once when service is first accessed
func (s *Service) InitService(ctx context.Context, projectID string) error {
	s.projectID = projectID
	client, err := NewClient(ctx)
	if err != nil {
		return err
	}
	s.client = client
	return nil
}

// Reinit reinitializes the service with a new project ID (on project switch)
func (s *Service) Reinit(ctx context.Context, projectID string) error {
	s.Reset()
	s.loaded = false // Force reload on next entry
	return s.InitService(ctx, projectID)
}

// Init returns startup commands (background tick)
func (s *Service) Init() tea.Cmd {
	return s.tick()
}

// tick creates a background ticker for cache invalidation/refresh
func (s *Service) tick() tea.Cmd {
	return tea.Tick(CacheTTL, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Refresh triggers a forced data reload with spinner
func (s *Service) Refresh() tea.Cmd {
	return tea.Batch(
		s.spinner.Start(""), // Empty string = playful random messages
		s.fetchDataCmd(true),
	)
}

// Reset clears the service state when navigating away
func (s *Service) Reset() {
	s.viewState = ViewList
	s.selectedItem = nil
	s.err = nil // CRITICAL: Always clear errors on reset
	s.pendingAction = ""
	s.table.SetCursor(0)
	s.filter.ExitFilterMode()
}

// IsRootView returns true if at the top-level list (used for 'q' navigation)
func (s *Service) IsRootView() bool {
	return s.viewState == ViewList
}

// Focus handles input focus - triggers initial load if needed
func (s *Service) Focus() {
	s.table.Focus()
}

// Blur handles loss of input focus
func (s *Service) Blur() {
	s.table.Blur()
}

// =============================================================================
// Update Loop
// =============================================================================

func (s *Service) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case components.SpinnerTickMsg:
		s.spinner, cmd = s.spinner.Update(msg)
		return s, cmd

	case tickMsg:
		return s, tea.Batch(s.fetchDataCmd(false), s.tick())

	case dataMsg:
		s.spinner.Stop()
		s.items = msg
		s.loaded = true
		s.filterSession.Apply(s.items)
		return s, func() tea.Msg { return core.LastUpdatedMsg(time.Now()) }

	case errMsg:
		s.spinner.Stop()
		s.err = msg
		return s, nil

	case actionResultMsg:
		s.spinner.Stop()
		if msg.err != nil {
			return s, func() tea.Msg {
				return core.ToastMsg{
					Message: fmt.Sprintf("Failed to %s %s: %v", msg.action, msg.name, msg.err),
					Type:    core.ToastError,
				}
			}
		}
		return s, tea.Batch(
			func() tea.Msg {
				return core.ToastMsg{
					Message: fmt.Sprintf("%s %s successfully", capitalize(msg.action), msg.name),
					Type:    core.ToastSuccess,
				}
			},
			s.Refresh(),
		)

	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
		s.table.HandleWindowSizeDefault(msg)

	case tea.MouseMsg:
		if s.viewState == ViewList {
			var updatedTable *components.StandardTable
			updatedTable, cmd = s.table.Update(msg)
			s.table = updatedTable
			return s, cmd
		}

	case tea.KeyMsg:
		return s.handleKeyMsg(msg)
	}

	return s, nil
}

// handleKeyMsg processes keyboard input based on current view state
func (s *Service) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if s.viewState == ViewList {
		result := s.filterSession.HandleKey(msg)
		if result.Handled {
			if result.Cmd != nil {
				return s, result.Cmd
			}
			if !result.ShouldContinue {
				return s, nil
			}
		}

		switch msg.String() {
		case "r":
			return s, s.Refresh()
		case "enter":
			items := s.getCurrentItems()
			if idx := s.table.Cursor(); idx >= 0 && idx < len(items) {
				s.selectedItem = &items[idx]
				s.viewState = ViewDetail
			}
			return s, nil
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
			s.selectedItem = nil
			return s, nil
		}
	}

	if s.viewState == ViewConfirmation {
		switch msg.String() {
		case "y", "enter":
			action := s.pendingAction
			s.viewState = s.actionSource
			s.pendingAction = ""
			return s, s.performActionCmd(action)
		case "n", "esc", "q":
			s.viewState = s.actionSource
			s.pendingAction = ""
			return s, nil
		}
	}

	return s, nil
}

// =============================================================================
// View Rendering
// =============================================================================

func (s *Service) View() string {
	if s.err != nil {
		return components.RenderError(s.err, s.Name(), "Builds")
	}

	if s.spinner.IsActive() {
		return s.spinner.View()
	}

	if s.viewState == ViewConfirmation && s.selectedItem != nil {
		return components.RenderConfirmation(s.pendingAction, s.selectedItem.Name, "build")
	}

	if s.viewState == ViewDetail {
		return s.renderDetailView()
	}

	return s.renderListView()
}

func (s *Service) renderListView() string {
	breadcrumb := components.Breadcrumb(
		fmt.Sprintf("Project: %s", s.projectID),
		s.Name(),
	)

	return lipgloss.JoinVertical(lipgloss.Left,
		breadcrumb,
		s.filter.View(),
		s.table.View(),
	)
}

func (s *Service) renderDetailView() string {
	if s.selectedItem == nil {
		return "Error: No item selected"
	}

	breadcrumb := components.Breadcrumb(
		fmt.Sprintf("Project: %s", s.projectID),
		s.Name(),
		s.selectedItem.Name,
	)

	card := components.DetailCard(components.DetailCardOpts{
		Title: "Build Details",
		Rows: []components.KeyValue{
			{Key: "Name", Value: s.selectedItem.Name},
			{Key: "Status", Value: s.selectedItem.Status},
			{Key: "Region", Value: s.selectedItem.Region},
			{Key: "ID", Value: s.selectedItem.ID},
		},
	})

	actions := components.RenderFooterHint("q Back")

	return lipgloss.JoinVertical(lipgloss.Left,
		breadcrumb,
		"",
		card,
		"",
		actions,
	)
}

// =============================================================================
// Data Fetching
// =============================================================================

func (s *Service) fetchDataCmd(force bool) tea.Cmd {
	return func() tea.Msg {
		cacheKey := fmt.Sprintf("cloudbuild_items:%s", s.projectID)

		if !force && s.cache != nil {
			if val, found := s.cache.Get(cacheKey); found {
				if items, ok := val.([]BuildItem); ok {
					return dataMsg(items)
				}
			}
		}

		if s.client == nil {
			return errMsg(fmt.Errorf("client not initialized"))
		}

		// Placeholder for now
		items := []BuildItem{}

		if s.cache != nil {
			s.cache.Set(cacheKey, items, CacheTTL)
		}

		return dataMsg(items)
	}
}

// =============================================================================
// Table Updates
// =============================================================================

func (s *Service) updateTable(items []BuildItem) {
	rows := make([]table.Row, len(items))
	for i, item := range items {
		rows[i] = table.Row{
			item.Name,
			components.RenderStatus(item.Status), 
			item.Region,
			item.ID,
		}
	}
	s.table.SetRows(rows)
}

func (s *Service) getCurrentItems() []BuildItem {
	return s.getFilteredItems(s.items, s.filter.Value())
}

func (s *Service) getFilteredItems(items []BuildItem, query string) []BuildItem {
	if query == "" {
		return items
	}
	return components.FilterSlice(items, query, func(item BuildItem, q string) bool {
		return components.ContainsMatch(item.Name, item.Status, item.Region, item.ID)(q)
	})
}

// =============================================================================
// Actions
// =============================================================================

func (s *Service) performActionCmd(action string) tea.Cmd {
	item := s.selectedItem
	if item == nil {
		return nil
	}

	return func() tea.Msg {
		if s.client == nil {
			return actionResultMsg{action: action, name: item.Name, err: fmt.Errorf("client not initialized")}
		}

		var err error
		switch action {
		default:
			err = fmt.Errorf("unknown action: %s", action)
		}

		return actionResultMsg{action: action, name: item.Name, err: err}
	}
}

// =============================================================================
// Helpers
// =============================================================================

func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return string(s[0]-32) + s[1:] 
}
