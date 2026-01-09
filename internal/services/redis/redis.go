package redis

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

const CacheTTL = 60 * time.Second

// -----------------------------------------------------------------------------
// Models
// -----------------------------------------------------------------------------

type tickMsg time.Time

type ViewState int

const (
	ViewList ViewState = iota
	ViewDetail
)

type instancesMsg []Instance
type errMsg error

// -----------------------------------------------------------------------------
// Service Definition
// -----------------------------------------------------------------------------

type Service struct {
	client    *Client
	projectID string
	table     table.Model

	filterInput textinput.Model
	filtering   bool

	instances []Instance
	loading   bool
	err       error

	viewState        ViewState
	selectedInstance *Instance

	cache *core.Cache
}

func NewService(cache *core.Cache) *Service {
	columns := []table.Column{
		{Title: "Name", Width: 30},
		{Title: "Location", Width: 15},
		{Title: "Tier", Width: 15},
		{Title: "Capacity", Width: 10},
		{Title: "State", Width: 10},
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
	ti.Placeholder = "Filter..."
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
	return "Memorystore (Redis)"
}

func (s *Service) ShortName() string {
	return "redis"
}

func (s *Service) HelpText() string {
	if s.viewState == ViewList {
		return "r:Refresh  /:Filter  Ent:Detail"
	}
	return "Esc/q:Back"
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
	return s.fetchInstancesCmd(true)
}

func (s *Service) Reset() {
	s.viewState = ViewList
	s.selectedInstance = nil
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
	st.Selected = lipgloss.NewStyle().Foreground(lipgloss.Color("229")).Background(lipgloss.Color("57"))
	s.table.SetStyles(st)
}

func (s *Service) Blur() {
	s.table.Blur()
	st := table.DefaultStyles()
	st.Header = styles.HeaderStyle
	st.Selected = lipgloss.NewStyle().Foreground(styles.ColorText).Background(lipgloss.Color("237"))
	s.table.SetStyles(st)
}

// -----------------------------------------------------------------------------
// Update
// -----------------------------------------------------------------------------

func (s *Service) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tickMsg:
		return s, tea.Batch(s.fetchInstancesCmd(false), s.tick())

	case instancesMsg:
		s.loading = false
		s.instances = msg
		s.updateTable(s.instances)
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
		if s.filtering {
			switch msg.String() {
			case "esc":
				s.filtering = false
				s.filterInput.Blur()
				s.filterInput.Reset()
				s.updateTable(s.instances)
				return s, nil
			case "enter":
				s.filtering = false
				s.filterInput.Blur()
				return s, nil
			}
			var inputCmd tea.Cmd
			s.filterInput, inputCmd = s.filterInput.Update(msg)
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
				if idx := s.table.Cursor(); idx >= 0 && idx < len(s.instances) {
					s.selectedInstance = &s.instances[idx]
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
				s.selectedInstance = nil
				return s, nil
			}
		}
	}
	return s, nil
}

// -----------------------------------------------------------------------------
// Data & Helpers
// -----------------------------------------------------------------------------

func (s *Service) fetchInstancesCmd(force bool) tea.Cmd {
	return func() tea.Msg {
		key := fmt.Sprintf("redis:%s", s.projectID)
		if !force && s.cache != nil {
			if val, found := s.cache.Get(key); found {
				if items, ok := val.([]Instance); ok {
					return instancesMsg(items)
				}
			}
		}
		if s.client == nil {
			return errMsg(fmt.Errorf("client not init"))
		}
		items, err := s.client.ListInstances(s.projectID)
		if err != nil {
			return errMsg(err)
		}
		if s.cache != nil {
			s.cache.Set(key, items, CacheTTL)
		}
		return instancesMsg(items)
	}
}

func (s *Service) updateTable(items []Instance) {
	rows := make([]table.Row, len(items))
	for i, item := range items {
		rows[i] = table.Row{
			item.Name,
			item.Location,
			item.Tier,
			fmt.Sprintf("%d GB", item.MemorySizeGb),
			item.State,
		}
	}
	s.table.SetRows(rows)
}
