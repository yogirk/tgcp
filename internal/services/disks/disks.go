package disks

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/yogirk/tgcp/internal/core"
	"github.com/yogirk/tgcp/internal/ui/components"
)

const CacheTTL = 30 * time.Second

// -----------------------------------------------------------------------------
// Models & Msgs
// -----------------------------------------------------------------------------

type tickMsg time.Time

type ViewState int

const (
	ViewList ViewState = iota
	ViewDetail
	ViewConfirmation
)

type disksMsg []Disk
type errMsg error
type actionResultMsg struct{ err error }

// -----------------------------------------------------------------------------
// Service Definition
// -----------------------------------------------------------------------------

type Service struct {
	client    *Client
	projectID string
	table     *components.StandardTable

	filter        components.FilterModel
	filterSession components.FilterSession[Disk]

	disks   []Disk
	spinner components.SpinnerModel
	err     error

	viewState    ViewState
	selectedDisk *Disk

	pendingAction string
	actionSource  ViewState

	cache *core.Cache
}

func NewService(cache *core.Cache) *Service {
	columns := []table.Column{
		{Title: "Name", Width: 25},
		{Title: "Zone", Width: 15},
		{Title: "Size", Width: 10},
		{Title: "Type", Width: 15},
		{Title: "Status", Width: 10},
		{Title: "Attached To", Width: 20},
	}

	t := components.NewStandardTable(columns)

	svc := &Service{
		table:     t,
		filter:    components.NewFilterWithPlaceholder("Filter disks..."),
		spinner:   components.NewSpinner(),
		viewState: ViewList,
		cache:     cache,
	}
	svc.filterSession = components.NewFilterSession(&svc.filter, svc.getFilteredDisks, svc.updateTable)
	return svc
}

func (s *Service) Name() string {
	return "GCE Disks"
}

func (s *Service) ShortName() string {
	return "disks"
}

func (s *Service) HelpText() string {
	if s.viewState == ViewList {
		return "r:Refresh  /:Filter  Ent:Detail"
	}
	if s.viewState == ViewDetail {
		return "Esc/q:Back  s:Snapshot"
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
		s.fetchDisksCmd(true),
	)
}

func (s *Service) Reset() {
	s.viewState = ViewList
	s.selectedDisk = nil
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
	case components.SpinnerTickMsg:
		s.spinner, cmd = s.spinner.Update(msg)
		return s, cmd

	case tickMsg:
		return s, tea.Batch(s.fetchDisksCmd(false), s.tick())

	case disksMsg:
		s.spinner.Stop()
		s.disks = msg
		s.filterSession.Apply(s.disks)
		return s, func() tea.Msg { return core.LastUpdatedMsg(time.Now()) }

	case errMsg:
		s.spinner.Stop()
		s.err = msg
		return s, nil

	case actionResultMsg:
		if msg.err != nil {
			s.err = msg.err
		}
		return s, nil

	case tea.WindowSizeMsg:
		s.table.HandleWindowSizeDefault(msg)

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

		if s.viewState == ViewList {
			switch msg.String() {
			case "r":
				return s, s.Refresh()
			case "enter":
				disks := s.getFilteredDisks(s.disks, s.filter.Value())
				if idx := s.table.Cursor(); idx >= 0 && idx < len(disks) {
					s.selectedDisk = &disks[idx]
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
				s.selectedDisk = nil
				return s, nil
			case "s":
				// Placeholder for Snapshot
				s.pendingAction = "snapshot"
				s.actionSource = ViewDetail
				s.viewState = ViewConfirmation
				return s, nil
			}
		}

		if s.viewState == ViewConfirmation {
			switch msg.String() {
			case "y", "enter":
				// No-op for MVP (Read-onlyish)
				// or implement snapshot call
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
// Views
// -----------------------------------------------------------------------------

func (s *Service) View() string {
	if s.err != nil {
		return components.RenderError(s.err, s.Name(), "Disks")
	}

	// Show spinner while loading
	if s.spinner.IsActive() {
		return s.spinner.View()
	}

	if s.viewState == ViewDetail {
		return s.renderDetailView()
	}
	if s.viewState == ViewConfirmation {
		if s.selectedDisk == nil {
			return "Error: No disk selected"
		}
		return components.RenderConfirmation("snapshot", s.selectedDisk.Name, "disk")
	}

	return s.renderListView()
}

func (s *Service) renderListView() string {
	// Filter Bar
	var content strings.Builder
	content.WriteString(components.Breadcrumb(
		fmt.Sprintf("Project %s", s.projectID),
		s.Name(),
		"Disks",
	))
	content.WriteString("\n")
	content.WriteString(s.filter.View())
	content.WriteString("\n")
	content.WriteString(s.table.View())
	return content.String()
}

// -----------------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------------

func (s *Service) fetchDisksCmd(force bool) tea.Cmd {
	return func() tea.Msg {
		key := fmt.Sprintf("disks:%s", s.projectID)

		if !force && s.cache != nil {
			if val, found := s.cache.Get(key); found {
				if items, ok := val.([]Disk); ok {
					return disksMsg(items)
				}
			}
		}

		if s.client == nil {
			return errMsg(fmt.Errorf("client not init"))
		}

		items, err := s.client.ListDisks(s.projectID)
		if err != nil {
			return errMsg(err)
		}

		if s.cache != nil {
			s.cache.Set(key, items, CacheTTL)
		}

		return disksMsg(items)
	}
}

func (s *Service) updateTable(items []Disk) {
	rows := make([]table.Row, len(items))
	for i, item := range items {
		attachedTo := "None"
		if len(item.Users) > 0 {
			// users link is like .../instances/instance-name
			parts := strings.Split(item.Users[0], "/")
			attachedTo = parts[len(parts)-1]
			if len(item.Users) > 1 {
				attachedTo += fmt.Sprintf(" (+%d)", len(item.Users)-1)
			}
		} else {
			attachedTo = "ORPHAN" // Will be plain text, but visually distinct by content
		}

		rows[i] = table.Row{
			item.Name,
			item.Zone,
			fmt.Sprintf("%d GB", item.SizeGb),
			item.ShortType(),
			item.Status,
			attachedTo,
		}
	}
	s.table.SetRows(rows)
}

// getFilteredDisks returns filtered disks based on the query string
func (s *Service) getFilteredDisks(disks []Disk, query string) []Disk {
	if query == "" {
		return disks
	}
	return components.FilterSlice(disks, query, func(disk Disk, q string) bool {
		return components.ContainsMatch(disk.Name, disk.Zone, disk.Type, disk.Status)(q)
	})
}
