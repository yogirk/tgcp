package secrets

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

const CacheTTL = 60 * time.Second // Secrets rarely change

// -----------------------------------------------------------------------------
// Message Types
// -----------------------------------------------------------------------------

type tickMsg time.Time
type secretsMsg []Secret
type versionsMsg []SecretVersion
type errMsg error

// ViewState defines the current UI state
type ViewState int

const (
	ViewList ViewState = iota
	ViewDetail
	ViewVersions
)

// -----------------------------------------------------------------------------
// Service Definition
// -----------------------------------------------------------------------------

// Service implements the services.Service interface for Secret Manager
type Service struct {
	client    *Client
	projectID string

	// Dimensions
	width  int
	height int

	// UI Components
	table         *components.StandardTable
	versionTable  *components.StandardTable
	filter        components.FilterModel
	filterSession components.FilterSession[Secret]
	spinner       components.SpinnerModel

	// Data State
	secrets  []Secret
	versions []SecretVersion
	err      error

	// View State
	viewState      ViewState
	selectedSecret *Secret

	// Cache
	cache *core.Cache
}

// NewService creates a new Secret Manager service
func NewService(cache *core.Cache) *Service {
	columns := []table.Column{
		{Title: "Name", Width: 35},
		{Title: "Replication", Width: 15},
		{Title: "Labels", Width: 25},
		{Title: "Created", Width: 20},
	}

	t := components.NewStandardTable(columns)

	versionColumns := []table.Column{
		{Title: "Version", Width: 10},
		{Title: "State", Width: 12},
		{Title: "Created", Width: 25},
	}
	vt := components.NewStandardTable(versionColumns)

	svc := &Service{
		table:        t,
		versionTable: vt,
		filter:       components.NewFilterWithPlaceholder("Filter secrets..."),
		spinner:      components.NewSpinner(),
		viewState:    ViewList,
		cache:        cache,
	}
	svc.filterSession = components.NewFilterSession(&svc.filter, svc.getFilteredSecrets, svc.updateTable)
	return svc
}

func (s *Service) Name() string {
	return "Secret Manager"
}

func (s *Service) ShortName() string {
	return "secrets"
}

func (s *Service) HelpText() string {
	switch s.viewState {
	case ViewList:
		return "r:Refresh  /:Filter  Enter:Detail"
	case ViewDetail:
		return "v:Versions  Esc/q:Back"
	case ViewVersions:
		return "Esc/q:Back to Detail"
	default:
		return ""
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
		s.fetchSecretsCmd(true),
	)
}

func (s *Service) Reset() {
	s.viewState = ViewList
	s.selectedSecret = nil
	s.versions = nil
	s.err = nil
	s.table.SetCursor(0)
	s.versionTable.SetCursor(0)
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
		return s, tea.Batch(s.fetchSecretsCmd(false), s.tick())

	case secretsMsg:
		s.spinner.Stop()
		s.secrets = msg
		s.filterSession.Apply(s.secrets)
		return s, func() tea.Msg { return core.LastUpdatedMsg(time.Now()) }

	case versionsMsg:
		s.spinner.Stop()
		s.versions = msg
		s.updateVersionTable(msg)
		return s, nil

	case errMsg:
		s.spinner.Stop()
		s.err = msg
		return s, nil

	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
		s.table.HandleWindowSizeDefault(msg)
		s.versionTable.HandleWindowSizeDefault(msg)

	case tea.KeyMsg:
		return s.handleKeyMsg(msg)
	}

	return s, nil
}

func (s *Service) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch s.viewState {
	case ViewList:
		// Handle filter
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
			secrets := s.getCurrentSecrets()
			if idx := s.table.Cursor(); idx >= 0 && idx < len(secrets) {
				s.selectedSecret = &secrets[idx]
				s.viewState = ViewDetail
			}
			return s, nil
		}

		var updatedTable *components.StandardTable
		updatedTable, cmd = s.table.Update(msg)
		s.table = updatedTable
		return s, cmd

	case ViewDetail:
		switch msg.String() {
		case "esc", "q":
			s.viewState = ViewList
			s.selectedSecret = nil
			s.versions = nil
			return s, nil
		case "v":
			// Fetch versions for selected secret
			if s.selectedSecret != nil {
				s.viewState = ViewVersions
				return s, tea.Batch(
					s.spinner.Start(""),
					s.fetchVersionsCmd(s.selectedSecret.FullName),
				)
			}
		}

	case ViewVersions:
		switch msg.String() {
		case "esc", "q":
			s.viewState = ViewDetail
			return s, nil
		}

		var updatedTable *components.StandardTable
		updatedTable, cmd = s.versionTable.Update(msg)
		s.versionTable = updatedTable
		return s, cmd
	}

	return s, nil
}

// -----------------------------------------------------------------------------
// View
// -----------------------------------------------------------------------------

func (s *Service) View() string {
	if s.err != nil {
		return components.RenderError(s.err, s.Name(), "Secrets")
	}

	if s.spinner.IsActive() {
		return s.spinner.View()
	}

	switch s.viewState {
	case ViewDetail:
		return s.renderDetailView()
	case ViewVersions:
		return s.renderVersionsView()
	default:
		return s.renderListView()
	}
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
	if s.selectedSecret == nil {
		return "No secret selected"
	}

	sec := s.selectedSecret

	breadcrumb := components.Breadcrumb(
		fmt.Sprintf("Project: %s", s.projectID),
		s.Name(),
		sec.Name,
	)

	// Format labels
	labelStr := "-"
	if len(sec.Labels) > 0 {
		var labels []string
		for k, v := range sec.Labels {
			labels = append(labels, fmt.Sprintf("%s=%s", k, v))
		}
		labelStr = fmt.Sprintf("%v", labels)
	}

	card := components.DetailCard(components.DetailCardOpts{
		Title: "Secret Details",
		Rows: []components.KeyValue{
			{Key: "Name", Value: sec.Name},
			{Key: "Replication", Value: sec.Replication},
			{Key: "Labels", Value: labelStr},
			{Key: "Created", Value: sec.CreateTime.Local().Format("2006-01-02 15:04:05")},
			{Key: "Resource Name", Value: sec.FullName},
		},
		FooterHint: "v Versions | q Back",
	})

	// Security note
	note := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Italic(true).
		Render("Note: Secret values are not displayed for security reasons.")

	return lipgloss.JoinVertical(lipgloss.Left,
		breadcrumb,
		"",
		card,
		"",
		note,
	)
}

func (s *Service) renderVersionsView() string {
	if s.selectedSecret == nil {
		return "No secret selected"
	}

	breadcrumb := components.Breadcrumb(
		fmt.Sprintf("Project: %s", s.projectID),
		s.Name(),
		s.selectedSecret.Name,
		"Versions",
	)

	return lipgloss.JoinVertical(lipgloss.Left,
		breadcrumb,
		"",
		s.versionTable.View(),
	)
}

// -----------------------------------------------------------------------------
// Data Fetching
// -----------------------------------------------------------------------------

func (s *Service) fetchSecretsCmd(force bool) tea.Cmd {
	return func() tea.Msg {
		cacheKey := fmt.Sprintf("secrets:%s", s.projectID)

		if !force && s.cache != nil {
			if val, found := s.cache.Get(cacheKey); found {
				if secrets, ok := val.([]Secret); ok {
					return secretsMsg(secrets)
				}
			}
		}

		if s.client == nil {
			return errMsg(fmt.Errorf("client not initialized"))
		}

		secrets, err := s.client.ListSecrets(s.projectID)
		if err != nil {
			return errMsg(err)
		}

		if s.cache != nil {
			s.cache.Set(cacheKey, secrets, CacheTTL)
		}

		return secretsMsg(secrets)
	}
}

func (s *Service) fetchVersionsCmd(secretName string) tea.Cmd {
	return func() tea.Msg {
		if s.client == nil {
			return errMsg(fmt.Errorf("client not initialized"))
		}

		versions, err := s.client.ListVersions(secretName)
		if err != nil {
			return errMsg(err)
		}

		return versionsMsg(versions)
	}
}

// -----------------------------------------------------------------------------
// Table Updates
// -----------------------------------------------------------------------------

func (s *Service) updateTable(secrets []Secret) {
	rows := make([]table.Row, len(secrets))
	for i, sec := range secrets {
		// Format labels for display
		labelStr := "-"
		if len(sec.Labels) > 0 {
			count := len(sec.Labels)
			if count == 1 {
				for k, v := range sec.Labels {
					labelStr = fmt.Sprintf("%s=%s", k, v)
				}
			} else {
				labelStr = fmt.Sprintf("%d labels", count)
			}
		}

		rows[i] = table.Row{
			sec.Name,
			sec.Replication,
			labelStr,
			sec.CreateTime.Local().Format("2006-01-02 15:04"),
		}
	}
	s.table.SetRows(rows)
}

func (s *Service) updateVersionTable(versions []SecretVersion) {
	rows := make([]table.Row, len(versions))
	for i, v := range versions {
		rows[i] = table.Row{
			v.Name,
			components.RenderStatus(v.State),
			v.CreateTime.Local().Format("2006-01-02 15:04:05"),
		}
	}
	s.versionTable.SetRows(rows)
}

func (s *Service) getCurrentSecrets() []Secret {
	return s.getFilteredSecrets(s.secrets, s.filter.Value())
}

func (s *Service) getFilteredSecrets(secrets []Secret, query string) []Secret {
	if query == "" {
		return secrets
	}
	return components.FilterSlice(secrets, query, func(sec Secret, q string) bool {
		// Build label string for search
		var labelStr string
		for k, v := range sec.Labels {
			labelStr += k + "=" + v + " "
		}
		return components.ContainsMatch(sec.Name, sec.Replication, labelStr)(q)
	})
}
