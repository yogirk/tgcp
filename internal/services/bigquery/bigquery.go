package bigquery

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yogirk/tgcp/internal/core"
	"github.com/yogirk/tgcp/internal/ui/components"
)

const CacheTTL = 5 * time.Minute

// -----------------------------------------------------------------------------
// Models
// -----------------------------------------------------------------------------

type tickMsg time.Time

type ViewState int

const (
	ViewDatasets ViewState = iota
	ViewTables
	ViewSchema
)

type datasetsMsg []Dataset
type tablesMsg []Table
type schemaMsg []SchemaField
type errMsg error

// -----------------------------------------------------------------------------
// Service Definition
// -----------------------------------------------------------------------------

type Service struct {
	client    *Client
	projectID string

	// Tables
	datasetTable *components.StandardTable
	tableTable   *components.StandardTable
	schemaTable  *components.StandardTable

	// UI
	filterInput textinput.Model

	// State
	datasets []Dataset
	tables   []Table
	schema   []SchemaField
	spinner  components.SpinnerModel
	err      error

	viewState       ViewState
	selectedDataset *Dataset
	selectedTable   *Table

	cache *core.Cache
}

func NewService(cache *core.Cache) *Service {
	// Dataset Table
	dsCols := []table.Column{
		{Title: "ID", Width: 30},
		{Title: "Location", Width: 15},
	}
	dsTable := components.NewStandardTable(dsCols)

	// Table Table
	tCols := []table.Column{
		{Title: "ID", Width: 30},
		{Title: "Type", Width: 10},
		{Title: "Rows", Width: 10},
		{Title: "Size", Width: 10},
	}
	tTable := components.NewStandardTable(tCols)

	// Schema Table
	sCols := []table.Column{
		{Title: "Field", Width: 20},
		{Title: "Type", Width: 15},
		{Title: "Mode", Width: 10},
		{Title: "Description", Width: 30},
	}
	sTable := components.NewStandardTable(sCols)

	return &Service{
		datasetTable: dsTable,
		tableTable:   tTable,
		schemaTable:  sTable,
		spinner:      components.NewSpinner(),
		viewState:    ViewDatasets,
		cache:        cache,
	}
}

func (s *Service) Name() string      { return "BigQuery" }
func (s *Service) ShortName() string { return "bq" }

func (s *Service) HelpText() string {
	if s.viewState == ViewDatasets {
		return "Ent:Select  r:Refresh"
	}
	if s.viewState == ViewTables {
		return "Ent:Schema  Esc/q:Back"
	}
	if s.viewState == ViewSchema {
		return "Esc/q:Back"
	}
	return ""
}

// -----------------------------------------------------------------------------
// Lifecycle
// -----------------------------------------------------------------------------

func (s *Service) InitService(ctx context.Context, projectID string) error {
	s.projectID = projectID
	client, err := NewClient(ctx, projectID)
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
	return tea.Tick(CacheTTL, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func (s *Service) Refresh() tea.Cmd {
	var fetchCmd tea.Cmd
	if s.viewState == ViewDatasets {
		fetchCmd = s.fetchDatasetsCmd()
	} else if s.viewState == ViewTables && s.selectedDataset != nil {
		fetchCmd = s.fetchTablesCmd()
	}
	if fetchCmd == nil {
		return nil
	}
	return tea.Batch(
		func() tea.Msg { return core.LoadingMsg{IsLoading: true} },
		fetchCmd,
		s.spinner.Start(""),
	)
}

func (s *Service) Reset() {
	s.viewState = ViewDatasets
	s.selectedDataset = nil
	s.selectedTable = nil
	s.err = nil
	s.datasetTable.SetCursor(0)
}

func (s *Service) IsRootView() bool {
	return s.viewState == ViewDatasets
}

func (s *Service) Focus() {
	s.datasetTable.Focus()
	s.tableTable.Focus()
	s.schemaTable.Focus()
}

func (s *Service) Blur() {
	s.datasetTable.Blur()
	s.tableTable.Blur()
	s.schemaTable.Blur()
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
		// Background refresh? Maybe only if at root.
		if s.viewState == ViewDatasets {
			return s, tea.Batch(s.fetchDatasetsCmd(), s.Init())
		}
		return s, s.Init()

	case datasetsMsg:
		s.spinner.Stop()
		s.datasets = msg
		s.updateDatasetTable()
		return s, tea.Batch(
			func() tea.Msg { return core.LoadingMsg{IsLoading: false} },
			func() tea.Msg { return core.LastUpdatedMsg(time.Now()) },
		)

	case tablesMsg:
		s.spinner.Stop()
		s.tables = msg
		s.updateTableTable()
		return s, tea.Batch(
			func() tea.Msg { return core.LoadingMsg{IsLoading: false} },
			func() tea.Msg { return core.LastUpdatedMsg(time.Now()) },
		)

	case schemaMsg:
		s.spinner.Stop()
		s.schema = msg
		s.updateSchemaTable()
		return s, tea.Batch(
			func() tea.Msg { return core.LoadingMsg{IsLoading: false} },
			func() tea.Msg { return core.LastUpdatedMsg(time.Now()) },
		)

	case errMsg:
		s.spinner.Stop()
		s.err = msg
		return s, func() tea.Msg { return core.LoadingMsg{IsLoading: false} }

	case tea.WindowSizeMsg:
		s.datasetTable.HandleWindowSizeDefault(msg)
		s.tableTable.HandleWindowSizeDefault(msg)
		s.schemaTable.HandleWindowSizeDefault(msg)

	case tea.KeyMsg:
		switch msg.String() {
		case "r":
			return s, s.Refresh()
		}

		if s.viewState == ViewDatasets {
			if msg.String() == "enter" {
				if s.datasetTable.Cursor() >= 0 && s.datasetTable.Cursor() < len(s.datasets) {
					s.selectedDataset = &s.datasets[s.datasetTable.Cursor()]
					s.viewState = ViewTables
					return s, tea.Batch(s.fetchTablesCmd(), s.spinner.Start(""))
				}
			}
			var updatedTable *components.StandardTable
			updatedTable, cmd = s.datasetTable.Update(msg)
			s.datasetTable = updatedTable
			return s, cmd
		}

		if s.viewState == ViewTables {
			if msg.String() == "esc" || msg.String() == "q" {
				s.viewState = ViewDatasets
				s.selectedDataset = nil
				return s, nil
			}
			if msg.String() == "enter" {
				if s.tableTable.Cursor() >= 0 && s.tableTable.Cursor() < len(s.tables) {
					s.selectedTable = &s.tables[s.tableTable.Cursor()]
					s.viewState = ViewSchema
					return s, tea.Batch(s.fetchSchemaCmd(), s.spinner.Start(""))
				}
			}
			var updatedTable *components.StandardTable
			updatedTable, cmd = s.tableTable.Update(msg)
			s.tableTable = updatedTable
			return s, cmd
		}

		if s.viewState == ViewSchema {
			if msg.String() == "esc" || msg.String() == "q" {
				s.viewState = ViewTables
				s.selectedTable = nil
				return s, nil
			}
			var updatedTable *components.StandardTable
			updatedTable, cmd = s.schemaTable.Update(msg)
			s.schemaTable = updatedTable
			return s, cmd
		}
	}
	return s, nil
}

// -----------------------------------------------------------------------------
// View
// -----------------------------------------------------------------------------

func (s *Service) View() string {
	if s.err != nil {
		return components.RenderError(s.err, s.Name(), "Datasets")
	}

	// Show spinner while loading
	if s.spinner.IsActive() {
		return s.spinner.View()
	}

	if s.viewState == ViewDatasets {
		breadcrumb := components.Breadcrumb(
			fmt.Sprintf("Project %s", s.projectID),
			s.Name(),
			"Datasets",
		)
		return lipgloss.JoinVertical(lipgloss.Left, breadcrumb, s.datasetTable.View())
	} else if s.viewState == ViewTables {
		header := components.Breadcrumb(
			fmt.Sprintf("Project %s", s.projectID),
			s.Name(),
			"Datasets",
			s.selectedDataset.ID,
			"Tables",
		)
		return lipgloss.JoinVertical(lipgloss.Left, header, s.tableTable.View())
	} else if s.viewState == ViewSchema {
		header := components.Breadcrumb(
			fmt.Sprintf("Project %s", s.projectID),
			s.Name(),
			"Datasets",
			s.selectedDataset.ID,
			"Tables",
			s.selectedTable.ID,
			"Schema",
		)
		return lipgloss.JoinVertical(lipgloss.Left, header, s.schemaTable.View())
	}
	return ""
}

// -----------------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------------

func (s *Service) fetchDatasetsCmd() tea.Cmd {
	return func() tea.Msg {
		if s.client == nil {
			return errMsg(fmt.Errorf("client not init"))
		}
		ds, err := s.client.ListDatasets(s.projectID)
		if err != nil {
			return errMsg(err)
		}
		return datasetsMsg(ds)
	}
}

func (s *Service) fetchTablesCmd() tea.Cmd {
	return func() tea.Msg {
		if s.selectedDataset == nil {
			return errMsg(fmt.Errorf("no dataset"))
		}
		tables, err := s.client.ListTables(s.selectedDataset.ID)
		if err != nil {
			return errMsg(err)
		}
		return tablesMsg(tables)
	}
}

func (s *Service) fetchSchemaCmd() tea.Cmd {
	return func() tea.Msg {
		if s.selectedTable == nil {
			return errMsg(fmt.Errorf("no table"))
		}
		fields, err := s.client.GetTableSchema(s.selectedDataset.ID, s.selectedTable.ID)
		if err != nil {
			return errMsg(err)
		}
		return schemaMsg(fields)
	}
}

func (s *Service) updateDatasetTable() {
	rows := make([]table.Row, len(s.datasets))
	for i, d := range s.datasets {
		rows[i] = table.Row{d.ID, d.Location}
	}
	s.datasetTable.SetRows(rows)
}

func (s *Service) updateTableTable() {
	rows := make([]table.Row, len(s.tables))
	for i, t := range s.tables {
		size := fmt.Sprintf("%d B", t.TotalBytes)
		if t.TotalBytes > 1024*1024*1024 {
			size = fmt.Sprintf("%.1f GB", float64(t.TotalBytes)/1024/1024/1024)
		} else if t.TotalBytes > 1024*1024 {
			size = fmt.Sprintf("%.1f MB", float64(t.TotalBytes)/1024/1024)
		}

		rows[i] = table.Row{t.ID, t.Type, fmt.Sprintf("%d", t.NumRows), size}
	}
	s.tableTable.SetRows(rows)
}

func (s *Service) updateSchemaTable() {
	rows := make([]table.Row, len(s.schema))
	for i, f := range s.schema {
		rows[i] = table.Row{f.Name, f.Type, f.Mode, f.Description}
	}
	s.schemaTable.SetRows(rows)
}
