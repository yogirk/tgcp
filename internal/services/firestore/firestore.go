package firestore

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

const CacheTTL = 120 * time.Second // Info changes rarely

// -----------------------------------------------------------------------------
// Models
// -----------------------------------------------------------------------------

type tickMsg time.Time

type ViewState int

const (
	ViewList ViewState = iota
	ViewDetail
	ViewNamespaces  // Datastore mode: list namespaces
	ViewKinds       // Datastore mode: list kinds in a namespace
)

type dbsMsg []Database
type namespacesMsg []Namespace
type kindsMsg []Kind
type errMsg error

// -----------------------------------------------------------------------------
// Service Definition
// -----------------------------------------------------------------------------

type Service struct {
	client    *Client
	projectID string
	table     *components.StandardTable

	filter        components.FilterModel
	filterSession components.FilterSession[Database]

	dbs     []Database
	spinner components.SpinnerModel
	err     error

	viewState  ViewState
	selectedDB *Database

	// Datastore mode navigation
	namespaces        []Namespace
	selectedNamespace *Namespace
	kinds             []Kind
	nsTable           *components.StandardTable
	kindTable         *components.StandardTable

	cache *core.Cache
}

func NewService(cache *core.Cache) *Service {
	columns := []table.Column{
		{Title: "Database ID", Width: 30},
		{Title: "Type", Width: 20},
		{Title: "Location", Width: 15},
		{Title: "Created", Width: 20},
	}

	t := components.NewStandardTable(columns)

	// Namespace table for Datastore mode
	nsColumns := []table.Column{
		{Title: "Namespace", Width: 50},
	}
	nsTable := components.NewStandardTable(nsColumns)

	// Kind table for Datastore mode
	kindColumns := []table.Column{
		{Title: "Kind", Width: 50},
	}
	kindTable := components.NewStandardTable(kindColumns)

	svc := &Service{
		table:     t,
		nsTable:   nsTable,
		kindTable: kindTable,
		filter:    components.NewFilterWithPlaceholder("Filter databases..."),
		spinner:   components.NewSpinner(),
		viewState: ViewList,
		cache:     cache,
	}
	svc.filterSession = components.NewFilterSession(&svc.filter, svc.getFilteredDBs, svc.updateTable)
	return svc
}

func (s *Service) Name() string {
	return "Firestore"
}

func (s *Service) ShortName() string {
	return "firestore"
}

func (s *Service) HelpText() string {
	switch s.viewState {
	case ViewList:
		return "r:Refresh  /:Filter  Ent:Detail"
	case ViewNamespaces:
		return "r:Refresh  Ent:Kinds  Esc/q:Back"
	case ViewKinds:
		return "Esc/q:Back"
	default:
		return "Esc/q:Back"
	}
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
		s.fetchDBsCmd(true),
	)
}

func (s *Service) Reset() {
	s.viewState = ViewList
	s.selectedDB = nil
	s.selectedNamespace = nil
	s.namespaces = nil
	s.kinds = nil
	s.err = nil
	s.table.SetCursor(0)
	s.nsTable.SetCursor(0)
	s.kindTable.SetCursor(0)
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
		return s, tea.Batch(s.fetchDBsCmd(false), s.tick())

	case dbsMsg:
		s.spinner.Stop()
		s.dbs = msg
		s.filterSession.Apply(s.dbs)
		return s, func() tea.Msg { return core.LastUpdatedMsg(time.Now()) }

	case namespacesMsg:
		s.spinner.Stop()
		s.namespaces = msg
		s.updateNsTable(msg)
		return s, nil

	case kindsMsg:
		s.spinner.Stop()
		s.kinds = msg
		s.updateKindTable(msg)
		return s, nil

	case errMsg:
		s.spinner.Stop()
		s.err = msg
		return s, nil

	case tea.WindowSizeMsg:
		s.table.HandleWindowSizeDefault(msg)
		s.nsTable.HandleWindowSizeDefault(msg)
		s.kindTable.HandleWindowSizeDefault(msg)

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
				dbs := s.getFilteredDBs(s.dbs, s.filter.Value())
				if idx := s.table.Cursor(); idx >= 0 && idx < len(dbs) {
					s.selectedDB = &dbs[idx]
					// For Datastore mode, go to namespaces; for Firestore mode, go to detail
					if s.selectedDB.IsDatastoreMode() {
						s.viewState = ViewNamespaces
						return s, tea.Batch(
							s.spinner.Start("Loading namespaces..."),
							s.fetchNamespacesCmd(),
						)
					}
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
				s.selectedDB = nil
				return s, nil
			}
		}

		if s.viewState == ViewNamespaces {
			switch msg.String() {
			case "esc", "q":
				s.viewState = ViewList
				s.selectedDB = nil
				s.namespaces = nil
				return s, nil
			case "r":
				return s, tea.Batch(
					s.spinner.Start("Loading namespaces..."),
					s.fetchNamespacesCmd(),
				)
			case "enter":
				if idx := s.nsTable.Cursor(); idx >= 0 && idx < len(s.namespaces) {
					s.selectedNamespace = &s.namespaces[idx]
					s.viewState = ViewKinds
					return s, tea.Batch(
						s.spinner.Start("Loading kinds..."),
						s.fetchKindsCmd(),
					)
				}
			}
			var updatedTable *components.StandardTable
			updatedTable, cmd = s.nsTable.Update(msg)
			s.nsTable = updatedTable
			return s, cmd
		}

		if s.viewState == ViewKinds {
			switch msg.String() {
			case "esc", "q":
				s.viewState = ViewNamespaces
				s.selectedNamespace = nil
				s.kinds = nil
				return s, nil
			}
			var updatedTable *components.StandardTable
			updatedTable, cmd = s.kindTable.Update(msg)
			s.kindTable = updatedTable
			return s, cmd
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

// getFilteredDBs returns filtered databases based on the query string
func (s *Service) getFilteredDBs(dbs []Database, query string) []Database {
	if query == "" {
		return dbs
	}
	return components.FilterSlice(dbs, query, func(db Database, q string) bool {
		return components.ContainsMatch(db.Name, db.Type, db.Location)(q)
	})
}

// fetchNamespacesCmd fetches namespaces for Datastore mode database
func (s *Service) fetchNamespacesCmd() tea.Cmd {
	return func() tea.Msg {
		if s.client == nil || s.selectedDB == nil {
			return errMsg(fmt.Errorf("client or database not selected"))
		}
		items, err := s.client.ListNamespaces(s.projectID, s.selectedDB.Name)
		if err != nil {
			return errMsg(err)
		}
		return namespacesMsg(items)
	}
}

// fetchKindsCmd fetches kinds for the selected namespace
func (s *Service) fetchKindsCmd() tea.Cmd {
	return func() tea.Msg {
		if s.client == nil || s.selectedDB == nil || s.selectedNamespace == nil {
			return errMsg(fmt.Errorf("client, database, or namespace not selected"))
		}
		items, err := s.client.ListKinds(s.projectID, s.selectedDB.Name, s.selectedNamespace.Name)
		if err != nil {
			return errMsg(err)
		}
		return kindsMsg(items)
	}
}

// updateNsTable updates the namespace table with items
func (s *Service) updateNsTable(items []Namespace) {
	rows := make([]table.Row, len(items))
	for i, item := range items {
		rows[i] = table.Row{item.Name}
	}
	s.nsTable.SetRows(rows)
}

// updateKindTable updates the kind table with items
func (s *Service) updateKindTable(items []Kind) {
	rows := make([]table.Row, len(items))
	for i, item := range items {
		rows[i] = table.Row{item.Name}
	}
	s.kindTable.SetRows(rows)
}
