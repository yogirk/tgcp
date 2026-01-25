package dataflow

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
// Models
// -----------------------------------------------------------------------------

type tickMsg time.Time

type ViewState int

const (
	ViewList ViewState = iota
	ViewDetail
)

type jobsMsg []Job
type errMsg error

// -----------------------------------------------------------------------------
// Service Definition
// -----------------------------------------------------------------------------

type Service struct {
	client    *Client
	projectID string
	table     *components.StandardTable

	filter        components.FilterModel
	filterSession components.FilterSession[Job]

	jobs    []Job
	spinner components.SpinnerModel
	err     error

	viewState   ViewState
	selectedJob *Job

	cache *core.Cache
}

func NewService(cache *core.Cache) *Service {
	columns := []table.Column{
		{Title: "Job Name", Width: 40},
		{Title: "Type", Width: 15},
		{Title: "State", Width: 15},
		{Title: "Location", Width: 10},
	}

	t := components.NewStandardTable(columns)

	svc := &Service{
		table:     t,
		filter:    components.NewFilterWithPlaceholder("Filter jobs..."),
		spinner:   components.NewSpinner(),
		viewState: ViewList,
		cache:     cache,
	}
	svc.filterSession = components.NewFilterSession(&svc.filter, svc.getFilteredJobs, svc.updateTable)
	return svc
}

func (s *Service) Name() string {
	return "Dataflow"
}

func (s *Service) ShortName() string {
	return "dataflow"
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
		s.fetchJobsCmd(true),
	)
}

func (s *Service) Reset() {
	s.viewState = ViewList
	s.selectedJob = nil
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
		return s, tea.Batch(s.fetchJobsCmd(false), s.tick())

	case jobsMsg:
		s.spinner.Stop()
		s.jobs = msg
		s.filterSession.Apply(s.jobs)
		return s, func() tea.Msg { return core.LastUpdatedMsg(time.Now()) }

	case errMsg:
		s.spinner.Stop()
		s.err = msg
		return s, nil

	case tea.WindowSizeMsg:
		s.table.HandleWindowSizeDefault(msg)

	case tea.MouseMsg:
		// Forward mouse events to table for click selection
		if s.viewState == ViewList {
			var updatedTable *components.StandardTable
			updatedTable, cmd = s.table.Update(msg)
			s.table = updatedTable
			return s, cmd
		}

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
				jobs := s.getFilteredJobs(s.jobs, s.filter.Value())
				if idx := s.table.Cursor(); idx >= 0 && idx < len(jobs) {
					s.selectedJob = &jobs[idx]
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
				s.selectedJob = nil
				return s, nil
			}
		}
	}
	return s, nil
}

// -----------------------------------------------------------------------------
// Data & Helpers
// -----------------------------------------------------------------------------

func (s *Service) fetchJobsCmd(force bool) tea.Cmd {
	return func() tea.Msg {
		key := fmt.Sprintf("dataflow:%s", s.projectID)
		if !force && s.cache != nil {
			if val, found := s.cache.Get(key); found {
				if items, ok := val.([]Job); ok {
					return jobsMsg(items)
				}
			}
		}
		if s.client == nil {
			return errMsg(fmt.Errorf("client not init"))
		}
		items, err := s.client.ListJobs(s.projectID)
		if err != nil {
			return errMsg(err)
		}
		if s.cache != nil {
			s.cache.Set(key, items, CacheTTL)
		}
		return jobsMsg(items)
	}
}

func (s *Service) updateTable(items []Job) {
	rows := make([]table.Row, len(items))
	for i, item := range items {
		cleanState := strings.Replace(item.State, "JOB_STATE_", "", 1)
		cleanType := strings.Replace(item.Type, "JOB_TYPE_", "", 1)

		rows[i] = table.Row{
			item.Name,
			cleanType,
			cleanState,
			item.Location,
		}
	}
	s.table.SetRows(rows)
}

// getFilteredJobs returns filtered jobs based on the query string
func (s *Service) getFilteredJobs(jobs []Job, query string) []Job {
	if query == "" {
		return jobs
	}
	return components.FilterSlice(jobs, query, func(job Job, q string) bool {
		return components.ContainsMatch(job.Name, job.Type, job.State, job.Location)(q)
	})
}
