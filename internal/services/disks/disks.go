package disks

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rk/tgcp/internal/core"
	"github.com/rk/tgcp/internal/styles"
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
	table     table.Model

	filterInput textinput.Model
	filtering   bool

	disks   []Disk
	loading bool
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

	ti := textinput.New()
	ti.Placeholder = "Filter disks..."
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
	return s.fetchDisksCmd(true)
}

func (s *Service) Reset() {
	s.viewState = ViewList
	s.selectedDisk = nil
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
// Update
// -----------------------------------------------------------------------------

func (s *Service) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tickMsg:
		return s, tea.Batch(s.fetchDisksCmd(false), s.tick())

	case disksMsg:
		s.loading = false
		s.disks = msg
		s.updateTable(s.disks)
		return s, func() tea.Msg { return core.LastUpdatedMsg(time.Now()) }

	case errMsg:
		s.loading = false
		s.err = msg

	case actionResultMsg:
		if msg.err != nil {
			s.err = msg.err
		}
		return s, nil

	case tea.WindowSizeMsg:
		const heightOffset = 6
		newHeight := msg.Height - heightOffset
		if newHeight < 5 {
			newHeight = 5
		}
		s.table.SetHeight(newHeight)

	case tea.KeyMsg:
		if s.filtering {
			switch msg.String() {
			case "esc":
				s.filtering = false
				s.filterInput.Blur()
				s.filterInput.Reset()
				s.updateTable(s.disks)
				return s, nil
			case "enter":
				s.filtering = false
				s.filterInput.Blur()
				return s, nil
			}
			var inputCmd tea.Cmd
			s.filterInput, inputCmd = s.filterInput.Update(msg)
			// TODO: Add filter logic if needed
			return s, inputCmd
		}

		if s.viewState == ViewList {
			switch msg.String() {
			case "/":
				s.filtering = true
				s.filterInput.Focus()
				return s, textinput.Blink
			case "r":
				return s, s.Refresh()
			case "enter":
				if idx := s.table.Cursor(); idx >= 0 && idx < len(s.disks) {
					s.selectedDisk = &s.disks[idx]
					s.viewState = ViewDetail
				}
			}
			s.table, cmd = s.table.Update(msg)
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
	if s.loading && len(s.disks) == 0 {
		return "Loading GCE Disks..."
	}
	if s.err != nil {
		return fmt.Sprintf("Error: %v", s.err)
	}

	if s.viewState == ViewDetail {
		return s.renderDetailView()
	}
	if s.viewState == ViewConfirmation {
		return fmt.Sprintf("Create snapshot of disk '%s'? (y/n)", s.selectedDisk.Name)
	}

	return s.renderListView()
}

func (s *Service) renderListView() string {
	return s.table.View()
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
