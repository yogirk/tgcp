package firestore

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

const CacheTTL = 120 * time.Second // Info changes rarely

// -----------------------------------------------------------------------------
// Models
// -----------------------------------------------------------------------------

type tickMsg time.Time

type ViewState int

const (
	ViewList ViewState = iota
	ViewDetail
)

type dbsMsg []Database
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

	dbs     []Database
	loading bool
	err     error

	viewState  ViewState
	selectedDB *Database

	cache *core.Cache
}

func NewService(cache *core.Cache) *Service {
	columns := []table.Column{
		{Title: "Database ID", Width: 30},
		{Title: "Type", Width: 20},
		{Title: "Location", Width: 15},
		{Title: "Created", Width: 20},
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
	ti.Placeholder = "Filter databases..."
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
	return "Firestore"
}

func (s *Service) ShortName() string {
	return "firestore"
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
	return s.fetchDBsCmd(true)
}

func (s *Service) Reset() {
	s.viewState = ViewList
	s.selectedDB = nil
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
		return s, tea.Batch(s.fetchDBsCmd(false), s.tick())

	case dbsMsg:
		s.loading = false
		s.dbs = msg
		s.updateTable(s.dbs)
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
				s.updateTable(s.dbs)
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
				if idx := s.table.Cursor(); idx >= 0 && idx < len(s.dbs) {
					s.selectedDB = &s.dbs[idx]
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
				s.selectedDB = nil
				return s, nil
			}
		}
	}
	return s, nil
}

// -----------------------------------------------------------------------------
// Data & Helpers
// -----------------------------------------------------------------------------

func (s *Service) fetchDBsCmd(force bool) tea.Cmd {
	return func() tea.Msg {
		key := fmt.Sprintf("firestore:%s", s.projectID)
		if !force && s.cache != nil {
			if val, found := s.cache.Get(key); found {
				if items, ok := val.([]Database); ok {
					return dbsMsg(items)
				}
			}
		}
		if s.client == nil {
			return errMsg(fmt.Errorf("client not init"))
		}
		items, err := s.client.ListDatabases(s.projectID)
		if err != nil {
			return errMsg(err)
		}
		if s.cache != nil {
			s.cache.Set(key, items, CacheTTL)
		}
		return dbsMsg(items)
	}
}

func (s *Service) updateTable(items []Database) {
	rows := make([]table.Row, len(items))
	for i, item := range items {
		cleanType := strings.Replace(item.Type, "FIRESTORE_", "", 1)
		rows[i] = table.Row{
			item.Name,
			cleanType,
			item.Location,
			item.CreateTime,
		}
	}
	s.table.SetRows(rows)
}
