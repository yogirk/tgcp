package gcs

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

const CacheTTL = 5 * time.Minute // Buckets don't change often

// -----------------------------------------------------------------------------
// Models
// -----------------------------------------------------------------------------

// Tick message for background refresh
type tickMsg time.Time

// ViewState defines the current UI state of the service
type ViewState int

const (
	ViewList ViewState = iota
	ViewDetail
	ViewObjects
	ViewConfirmation
)

// bucketsMsg is the message used to pass fetched data
type bucketsMsg []Bucket

type objectsMsg []Object

// errMsg is the standard error message
type errMsg error

// -----------------------------------------------------------------------------
// Service Definition
// -----------------------------------------------------------------------------

// Service implements the services.Service interface
type Service struct {
	client      *Client
	projectID   string
	table       table.Model
	objectTable table.Model

	// UI Components
	filterInput textinput.Model
	filtering   bool

	// State
	buckets []Bucket
	objects []Object
	loading bool
	err     error

	// View State
	viewState      ViewState
	selectedBucket *Bucket
	selectedObject *Object
	currentPrefix  string

	// Cache
	cache *core.Cache
}

// NewService creates a new instance of the service
func NewService(cache *core.Cache) *Service {
	// 1. Table Setup
	columns := []table.Column{
		{Title: "Name", Width: 35},
		{Title: "Location", Width: 15},
		{Title: "Class", Width: 15},
		{Title: "Created", Width: 20},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	// Apply styles
	s := table.DefaultStyles()
	s.Header = styles.HeaderStyle
	s.Selected = lipgloss.NewStyle().
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	// 1b. Object Table Setup
	objColumns := []table.Column{
		{Title: "Name", Width: 40},
		{Title: "Type", Width: 15},
		{Title: "Size", Width: 10},
		{Title: "Updated", Width: 15},
	}
	ot := table.New(
		table.WithColumns(objColumns),
		table.WithFocused(true),
		table.WithHeight(10),
	)
	ot.SetStyles(s)

	// 2. Filter Input Setup
	ti := textinput.New()
	ti.Placeholder = "Filter buckets..."
	ti.Prompt = "/ "
	ti.CharLimit = 100
	ti.Width = 50

	return &Service{
		table:       t,
		objectTable: ot,
		filterInput: ti,
		viewState:   ViewList,
		cache:       cache,
	}
}

// Name returns the full human-readable name
func (s *Service) Name() string {
	return "Cloud Storage"
}

// ShortName returns the specialized identifier
func (s *Service) ShortName() string {
	return "gcs"
}

// HelpText returns context-aware keybindings
func (s *Service) HelpText() string {
	if s.viewState == ViewList {
		return "r:Refresh  /:Filter  Ent:Detail"
	}
	if s.viewState == ViewDetail {
		return "Esc/q:Back"
	}
	if s.viewState == ViewObjects {
		return "Enter:Open  Esc/q:Back/Up"
	}
	return ""
}

// -----------------------------------------------------------------------------
// Lifecycle & Interface Implementation
// -----------------------------------------------------------------------------

// InitService initializes the API client
func (s *Service) InitService(ctx context.Context, projectID string) error {
	s.projectID = projectID
	client, err := NewClient(ctx)
	if err != nil {
		return err
	}
	s.client = client
	return nil
}

// Init startup commands
func (s *Service) Init() tea.Cmd {
	return s.tick()
}

// tick creates a background ticker for cache invalidation
func (s *Service) tick() tea.Cmd {
	return tea.Tick(CacheTTL, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Refresh triggers a forced data reload
func (s *Service) Refresh() tea.Cmd {
	s.loading = true
	return s.fetchBucketsCmd(true)
}

// Reset clears the service state when navigating away or switching projects
func (s *Service) Reset() {
	s.viewState = ViewList
	s.selectedBucket = nil
	s.err = nil          // CRITICAL: Always clear errors on reset
	s.table.SetCursor(0) // Reset table position
	s.currentPrefix = ""
	s.selectedObject = nil
}

// IsRootView returns true if we are at the top-level list
func (s *Service) IsRootView() bool {
	return s.viewState == ViewList
}

// Focus handles input focus (Visual Highlight)
func (s *Service) Focus() {
	s.table.Focus()
	st := table.DefaultStyles()
	st.Header = styles.HeaderStyle
	st.Selected = lipgloss.NewStyle().
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	s.table.SetStyles(st)
	s.objectTable.SetStyles(st)
}

// Blur handles loss of input focus (Visual Dimming)
func (s *Service) Blur() {
	s.table.Blur()
	st := table.DefaultStyles()
	st.Header = styles.HeaderStyle
	st.Selected = lipgloss.NewStyle().
		Foreground(styles.ColorText).
		Background(lipgloss.Color("237")). // Dark grey
		Bold(false)
	s.table.SetStyles(st)
	s.objectTable.SetStyles(st)
}

// -----------------------------------------------------------------------------
// Update Loop
// -----------------------------------------------------------------------------

func (s *Service) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	// 1. Background Tick
	case tickMsg:
		return s, tea.Batch(s.fetchBucketsCmd(false), s.tick())

	// 2. Data Loaded
	case bucketsMsg:
		s.loading = false
		s.buckets = msg
		s.updateTable(s.buckets)
		return s, func() tea.Msg { return core.LastUpdatedMsg(time.Now()) }

	case objectsMsg:
		s.loading = false
		s.objects = msg
		s.updateObjectTable(s.objects)
		s.objectTable.SetCursor(0)
		return s, func() tea.Msg { return core.LastUpdatedMsg(time.Now()) }

	// 3. Error Handling
	case errMsg:
		s.loading = false
		s.err = msg

	// 4. Window Resize
	case tea.WindowSizeMsg:
		const heightOffset = 6
		newHeight := msg.Height - heightOffset
		if newHeight < 5 {
			newHeight = 5
		}
		s.table.SetHeight(newHeight)
		s.objectTable.SetHeight(newHeight)

	// 5. User Input
	case tea.KeyMsg:
		if s.viewState == ViewList {
			switch msg.String() {
			case "r":
				return s, s.Refresh()
			case "enter":
				// Handle bucket selection
				if s.selectedBucket == nil {
					buckets := s.buckets
					if idx := s.table.Cursor(); idx >= 0 && idx < len(buckets) {
						s.selectedBucket = &buckets[idx]
						// Switch to Object View
						s.viewState = ViewObjects
						s.currentPrefix = ""
						s.loading = true
						return s, s.fetchObjectsCmd()
						// s.viewState = ViewDetail // Not fully implemented for objects yet
					}
				}
			case "l": // Logs
				buckets := s.buckets
				if idx := s.table.Cursor(); idx >= 0 && idx < len(buckets) {
					b := buckets[idx]
					filter := fmt.Sprintf(`resource.type="gcs_bucket" AND resource.labels.bucket_name="%s"`, b.Name)
					return s, func() tea.Msg { return core.SwitchToLogsMsg{Filter: filter} }
				}
			}
			s.table, cmd = s.table.Update(msg)
			return s, cmd

		} else if s.viewState == ViewObjects {
			switch msg.String() {
			case "esc", "q":
				if s.currentPrefix == "" {
					s.viewState = ViewList
					s.selectedBucket = nil
				} else {
					s.currentPrefix = parentPrefix(s.currentPrefix)
					s.loading = true
					return s, s.fetchObjectsCmd()
				}
				return s, nil
			case "enter":
				// Drill down or select
				objs := s.objects
				if idx := s.objectTable.Cursor(); idx >= 0 && idx < len(objs) {
					obj := objs[idx]
					if obj.Type == "Folder" {
						s.currentPrefix = obj.Name
						s.loading = true
						return s, s.fetchObjectsCmd()
					} else {
						// Open Object Details?
						// For now, no-op or maybe ViewDetail if we implemented it for objects
						s.selectedObject = &obj
						// s.viewState = ViewDetail // Not fully implemented for objects yet
					}
				}
			}
			s.objectTable, cmd = s.objectTable.Update(msg)
			return s, cmd

		} else if s.viewState == ViewDetail {
			switch msg.String() {
			case "q", "esc":
				s.viewState = ViewList
				s.selectedBucket = nil
				return s, nil
			}
		}
	}

	return s, nil
}

// -----------------------------------------------------------------------------
// View Rendering
// -----------------------------------------------------------------------------

func (s *Service) View() string {
	if s.loading {
		return "Loading storage buckets..."
	}
	if s.err != nil {
		return fmt.Sprintf("Error fetching buckets: %v", s.err)
	}

	if s.viewState == ViewDetail {
		return s.renderDetailView()
	}

	if s.viewState == ViewObjects {
		return s.renderObjectListView()
	}

	// Default: List View
	return s.renderListView()
}

func (s *Service) renderListView() string {
	// Main table view
	return s.table.View()
}

func (s *Service) renderObjectListView() string {
	header := styles.SubtleStyle.Render(fmt.Sprintf("Cloud Storage > %s > %s", s.selectedBucket.Name, s.currentPrefix))
	return lipgloss.JoinVertical(lipgloss.Left, header, s.objectTable.View())
}

func (s *Service) renderDetailView() string {
	if s.selectedBucket == nil {
		return "No bucket selected"
	}

	b := s.selectedBucket

	// Title
	// Title
	title := styles.SubtleStyle.Render(fmt.Sprintf("Cloud Storage > %s", b.Name))

	// Details
	details := fmt.Sprintf(`
%s %s
%s %s
%s %s
`,
		styles.LabelStyle.Render("Location:"), styles.ValueStyle.Render(b.Location),
		styles.LabelStyle.Render("Class:"), styles.ValueStyle.Render(b.StorageClass),
		styles.LabelStyle.Render("Created:"), styles.ValueStyle.Render(b.Created.Format(time.RFC822)),
	)

	// Wrap in a box
	content := lipgloss.JoinVertical(lipgloss.Left,
		title,
		details,
		"",
		styles.SubtleStyle.Render("Press 'q' or 'esc' to return"),
	)

	return styles.FocusedBoxStyle.Render(content)
}

// -----------------------------------------------------------------------------
// Helper Commands
// -----------------------------------------------------------------------------

func (s *Service) fetchBucketsCmd(force bool) tea.Cmd {
	return func() tea.Msg {
		key := "gcs_buckets"

		// 1. Check Cache
		if !force && s.cache != nil {
			if val, found := s.cache.Get(key); found {
				if buckets, ok := val.([]Bucket); ok {
					return bucketsMsg(buckets)
				}
			}
		}

		// 2. API Call
		if s.client == nil {
			return errMsg(fmt.Errorf("client not initialized"))
		}
		buckets, err := s.client.ListBuckets(s.projectID)
		if err != nil {
			return errMsg(err)
		}

		// 3. Update Cache
		if s.cache != nil {
			s.cache.Set(key, buckets, CacheTTL)
		}

		return bucketsMsg(buckets)
	}
}

func (s *Service) updateTable(items []Bucket) {
	rows := make([]table.Row, len(items))
	for i, item := range items {
		rows[i] = table.Row{
			item.Name,
			item.Location,
			item.StorageClass,
			item.Created.Format("2006-01-02"),
		}
	}
	s.table.SetRows(rows)
}

func (s *Service) fetchObjectsCmd() tea.Cmd {
	return func() tea.Msg {
		// No caching for object browsing for now to keep it simple and fresh
		if s.selectedBucket == nil {
			return errMsg(fmt.Errorf("no bucket selected"))
		}

		objs, err := s.client.ListObjects(s.selectedBucket.Name, s.currentPrefix)
		if err != nil {
			return errMsg(err)
		}
		return objectsMsg(objs)
	}
}

func (s *Service) updateObjectTable(items []Object) {
	rows := make([]table.Row, len(items))
	for i, item := range items {
		sizeStr := fmt.Sprintf("%d B", item.Size)
		if item.Type == "Folder" {
			sizeStr = "-"
		} else if item.Size > 1024*1024 {
			sizeStr = fmt.Sprintf("%.1f MB", float64(item.Size)/1024/1024)
		} else if item.Size > 1024 {
			sizeStr = fmt.Sprintf("%.1f KB", float64(item.Size)/1024)
		}

		rows[i] = table.Row{
			item.Name,
			item.Type,
			sizeStr,
			item.Updated.Format("Jan 02 15:04"),
		}
	}
	s.objectTable.SetRows(rows)
}

// parentPrefix calculates the parent folder prefix
func parentPrefix(p string) string {
	if p == "" {
		return ""
	}
	// Remove trailing slash if exists (folders usually have it)
	if p[len(p)-1] == '/' {
		p = p[:len(p)-1]
	}
	// Find last slash
	for i := len(p) - 1; i >= 0; i-- {
		if p[i] == '/' {
			return p[:i+1]
		}
	}
	return ""
}
