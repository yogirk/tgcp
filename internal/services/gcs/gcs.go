package gcs

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yogirk/tgcp/internal/core"
	"github.com/yogirk/tgcp/internal/ui/components"
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
	table       *components.StandardTable
	objectTable *components.StandardTable

	// UI Components
	filter              components.FilterModel
	bucketFilterSession components.FilterSession[Bucket]
	objectFilterSession components.FilterSession[Object]
	spinner             components.SpinnerModel

	// State
	buckets []Bucket
	objects []Object
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

	t := components.NewStandardTable(columns)

	// 1b. Object Table Setup
	objColumns := []table.Column{
		{Title: "Name", Width: 40},
		{Title: "Type", Width: 15},
		{Title: "Size", Width: 10},
		{Title: "Updated", Width: 15},
	}
	ot := components.NewStandardTable(objColumns)

	svc := &Service{
		table:       t,
		objectTable: ot,
		filter:      components.NewFilterWithPlaceholder("Filter buckets..."),
		spinner:     components.NewSpinner(),
		viewState:   ViewList,
		cache:       cache,
	}
	svc.bucketFilterSession = components.NewFilterSession(&svc.filter, svc.getFilteredBuckets, svc.updateTable)
	svc.objectFilterSession = components.NewFilterSession(&svc.filter, svc.getFilteredObjects, svc.updateObjectTable)
	return svc
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
		return "Enter:Browse Objects  Esc/q:Back"
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

// Reinit reinitializes the service with a new project ID
func (s *Service) Reinit(ctx context.Context, projectID string) error {
	s.Reset()
	return s.InitService(ctx, projectID)
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
	return tea.Batch(
		s.spinner.Start(""), // Start animated spinner
		s.fetchBucketsCmd(true),
	)
}

// Reset clears the service state when navigating away or switching projects
func (s *Service) Reset() {
	s.viewState = ViewList
	s.selectedBucket = nil
	s.err = nil          // CRITICAL: Always clear errors on reset
	s.table.SetCursor(0) // Reset table position
	s.currentPrefix = ""
	s.selectedObject = nil
	s.filter.ExitFilterMode()
}

// IsRootView returns true if we are at the top-level list
func (s *Service) IsRootView() bool {
	return s.viewState == ViewList
}

// Focus handles input focus (Visual Highlight)
func (s *Service) Focus() {
	s.table.Focus()
	s.objectTable.Focus()
}

// Blur handles loss of input focus (Visual Dimming)
func (s *Service) Blur() {
	s.table.Blur()
	s.objectTable.Blur()
}

// -----------------------------------------------------------------------------
// Update Loop
// -----------------------------------------------------------------------------

func (s *Service) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	// Spinner Animation
	case components.SpinnerTickMsg:
		s.spinner, cmd = s.spinner.Update(msg)
		return s, cmd

	// 1. Background Tick
	case tickMsg:
		return s, tea.Batch(s.fetchBucketsCmd(false), s.tick())

	// 2. Data Loaded
	case bucketsMsg:
		s.spinner.Stop()
		s.buckets = msg
		s.bucketFilterSession.Apply(s.buckets)
		return s, func() tea.Msg { return core.LastUpdatedMsg(time.Now()) }

	case objectsMsg:
		s.spinner.Stop()
		s.objects = msg
		s.objectFilterSession.Apply(s.objects)
		s.objectTable.SetCursor(0)
		return s, func() tea.Msg { return core.LastUpdatedMsg(time.Now()) }

	// 3. Error Handling
	case errMsg:
		s.spinner.Stop()
		s.err = msg
		return s, nil

	// 4. Window Resize
	case tea.WindowSizeMsg:
		s.table.HandleWindowSizeDefault(msg)
		s.objectTable.HandleWindowSizeDefault(msg)

	// 4.5 Mouse Input
	case tea.MouseMsg:
		// Forward mouse events to active table for click selection
		if s.viewState == ViewList {
			var updatedTable *components.StandardTable
			updatedTable, cmd = s.table.Update(msg)
			s.table = updatedTable
			return s, cmd
		} else if s.viewState == ViewObjects {
			var updatedTable *components.StandardTable
			updatedTable, cmd = s.objectTable.Update(msg)
			s.objectTable = updatedTable
			return s, cmd
		}

	// 5. User Input
	case tea.KeyMsg:
		// Handle filter mode (list or object view)
		if s.viewState == ViewList || s.viewState == ViewObjects {
			var result components.FilterUpdateResult
			if s.viewState == ViewList {
				result = s.bucketFilterSession.HandleKey(msg)
			} else {
				result = s.objectFilterSession.HandleKey(msg)
			}

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
				// Handle bucket selection -> Go to Details
				if s.selectedBucket == nil {
					buckets := s.getFilteredBuckets(s.buckets, s.filter.Value())
					if idx := s.table.Cursor(); idx >= 0 && idx < len(buckets) {
						s.selectedBucket = &buckets[idx]
						s.viewState = ViewDetail
					}
				}
			}
			var updatedTable *components.StandardTable
			updatedTable, cmd = s.table.Update(msg)
			s.table = updatedTable
			return s, cmd

		} else if s.viewState == ViewDetail {
			switch msg.String() {
			case "q", "esc":
				s.viewState = ViewList
				s.selectedBucket = nil
				s.bucketFilterSession.Apply(s.buckets)
				return s, nil
			case "enter":
				// Go to Object Browser
				s.viewState = ViewObjects
				s.currentPrefix = ""
				return s, tea.Batch(s.fetchObjectsCmd(), s.spinner.Start(""))
			}

		} else if s.viewState == ViewObjects {
			switch msg.String() {
			case "esc", "q":
				if s.currentPrefix == "" {
					s.viewState = ViewDetail // Back to Details
				} else {
					s.currentPrefix = parentPrefix(s.currentPrefix)
					return s, tea.Batch(s.fetchObjectsCmd(), s.spinner.Start(""))
				}
				return s, nil
			case "enter":
				// Drill down or select
				objs := s.objects
				if idx := s.objectTable.Cursor(); idx >= 0 && idx < len(objs) {
					obj := objs[idx]
					if obj.Type == "Folder" {
						s.currentPrefix = obj.Name
						return s, tea.Batch(s.fetchObjectsCmd(), s.spinner.Start(""))
					} else {
						s.selectedObject = &obj
					}
				}
			}
			var updatedTable *components.StandardTable
			updatedTable, cmd = s.objectTable.Update(msg)
			s.objectTable = updatedTable
			return s, cmd
		}
	}

	return s, nil
}

// -----------------------------------------------------------------------------
// View Rendering
// -----------------------------------------------------------------------------

func (s *Service) View() string {
	if s.err != nil {
		return components.RenderError(s.err, s.Name(), "Buckets")
	}

	// Show animated spinner while loading
	if s.spinner.IsActive() {
		return s.spinner.View()
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
	// Filter Bar
	var content strings.Builder
	content.WriteString(components.Breadcrumb(
		fmt.Sprintf("Project %s", s.projectID),
		s.Name(),
		"Buckets",
	))
	content.WriteString("\n")
	content.WriteString(s.filter.View())
	content.WriteString("\n")
	content.WriteString(s.table.View())
	return content.String()
}

func (s *Service) renderObjectListView() string {
	prefix := s.currentPrefix
	if prefix == "" {
		prefix = "/"
	}
	header := components.Breadcrumb(
		fmt.Sprintf("Project %s", s.projectID),
		s.Name(),
		"Buckets",
		s.selectedBucket.Name,
		prefix,
	)
	return lipgloss.JoinVertical(lipgloss.Left, header, s.filter.View(), s.objectTable.View())
}

func (s *Service) renderDetailView() string {
	if s.selectedBucket == nil {
		return "No bucket selected"
	}

	b := s.selectedBucket

	title := components.Breadcrumb(
		fmt.Sprintf("Project %s", s.projectID),
		s.Name(),
		"Buckets",
		b.Name,
	)

	view := components.DetailCard(components.DetailCardOpts{
		Title: "Bucket Details",
		Rows: []components.KeyValue{
			{Key: "Name", Value: b.Name},
			{Key: "Location", Value: b.Location},
			{Key: "Class", Value: b.StorageClass},
			{Key: "Created", Value: b.Created.Format(time.RFC822)},
		},
	})

	// Action Bar
	actions := components.RenderFooterHint("Enter Browse Objects | q Back")

	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		"\n",
		view,
		"\n",
		actions,
	)
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

// getFilteredBuckets returns filtered buckets based on the query string
func (s *Service) getFilteredBuckets(buckets []Bucket, query string) []Bucket {
	if query == "" {
		return buckets
	}
	return components.FilterSlice(buckets, query, func(bucket Bucket, q string) bool {
		return components.ContainsMatch(bucket.Name, bucket.Location, bucket.StorageClass)(q)
	})
}

func (s *Service) getFilteredObjects(objects []Object, query string) []Object {
	if query == "" {
		return objects
	}
	return components.FilterSlice(objects, query, func(obj Object, q string) bool {
		return components.ContainsMatch(obj.Name, obj.Type)(q)
	})
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
