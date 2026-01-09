package pubsub

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
	ViewListTopics ViewState = iota // Default
	ViewListSubs
	ViewDetailTopic
	ViewDetailSub
)

type topicsMsg []Topic
type subsMsg []Subscription
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

	topics []Topic
	subs   []Subscription

	loading bool
	err     error

	viewState     ViewState
	selectedTopic *Topic
	selectedSub   *Subscription

	cache *core.Cache
}

func NewService(cache *core.Cache) *Service {
	// Default to Topic Columns
	columns := []table.Column{
		{Title: "Topic Name", Width: 40},
		{Title: "KMS Key", Width: 30},
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
		viewState:   ViewListTopics,
		cache:       cache,
	}
}

func (s *Service) Name() string {
	return "Pub/Sub"
}

func (s *Service) ShortName() string {
	return "pubsub"
}

func (s *Service) HelpText() string {
	if s.viewState == ViewListTopics {
		return "r:Refresh  /:Filter  s:Switch to Subs  Ent:Detail"
	}
	if s.viewState == ViewListSubs {
		return "r:Refresh  /:Filter  t:Switch to Topics  Ent:Detail"
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
	// Refresh both? Or just current view?
	// Simpler to refresh both
	return tea.Batch(s.fetchTopicsCmd(true), s.fetchSubsCmd(true))
}

func (s *Service) Reset() {
	s.viewState = ViewListTopics
	s.selectedTopic = nil
	s.selectedSub = nil
	s.err = nil
	s.table.SetCursor(0)
	s.filtering = false
	s.filterInput.Reset()
	s.updateTopicTable(s.topics) // Reset cols
}

func (s *Service) IsRootView() bool {
	return s.viewState == ViewListTopics || s.viewState == ViewListSubs
}

func (s *Service) Focus() {
	s.table.Focus()
	// Style reset
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
		return s, tea.Batch(s.fetchTopicsCmd(false), s.fetchSubsCmd(false), s.tick())

	case topicsMsg:
		s.topics = msg
		if s.viewState == ViewListTopics {
			s.loading = false
			s.updateTopicTable(s.topics)
		}
		return s, func() tea.Msg { return core.LastUpdatedMsg(time.Now()) }

	case subsMsg:
		s.subs = msg
		if s.viewState == ViewListSubs {
			s.loading = false
			s.updateSubTable(s.subs)
		}
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
				if s.viewState == ViewListTopics {
					s.updateTopicTable(s.topics)
				} else {
					s.updateSubTable(s.subs)
				}
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

		// Root Views (Topics or Subs)
		if s.viewState == ViewListTopics || s.viewState == ViewListSubs {
			switch msg.String() {
			case "/":
				s.filtering = true
				s.filterInput.Focus()
				return s, textinput.Blink
			case "r":
				return s, s.Refresh()
			case "s": // Switch to Subs
				if s.viewState == ViewListTopics {
					s.viewState = ViewListSubs
					s.updateSubTable(s.subs) // Render existing if available
					if len(s.subs) == 0 {
						s.loading = true
						return s, s.fetchSubsCmd(true)
					}
				}
			case "t": // Switch to Topics
				if s.viewState == ViewListSubs {
					s.viewState = ViewListTopics
					s.updateTopicTable(s.topics)
					if len(s.topics) == 0 {
						s.loading = true
						return s, s.fetchTopicsCmd(true)
					}
				}
			case "enter":
				if s.viewState == ViewListTopics {
					if idx := s.table.Cursor(); idx >= 0 && idx < len(s.topics) {
						s.selectedTopic = &s.topics[idx]
						s.viewState = ViewDetailTopic
					}
				} else {
					if idx := s.table.Cursor(); idx >= 0 && idx < len(s.subs) {
						s.selectedSub = &s.subs[idx]
						s.viewState = ViewDetailSub
					}
				}
			}
			s.table, cmd = s.table.Update(msg)
			return s, cmd
		}

		// Detail Views
		if s.viewState == ViewDetailTopic || s.viewState == ViewDetailSub {
			switch msg.String() {
			case "esc", "q":
				if s.viewState == ViewDetailTopic {
					s.viewState = ViewListTopics
					s.selectedTopic = nil
				} else {
					s.viewState = ViewListSubs
					s.selectedSub = nil
				}
				return s, nil
			}
		}
	}

	return s, nil
}

// -----------------------------------------------------------------------------
// Data Fetching
// -----------------------------------------------------------------------------

func (s *Service) fetchTopicsCmd(force bool) tea.Cmd {
	return func() tea.Msg {
		key := fmt.Sprintf("pubsub_topics:%s", s.projectID)
		if !force && s.cache != nil {
			if val, found := s.cache.Get(key); found {
				if items, ok := val.([]Topic); ok {
					return topicsMsg(items)
				}
			}
		}
		if s.client == nil {
			return errMsg(fmt.Errorf("client not init"))
		}
		items, err := s.client.ListTopics(s.projectID)
		if err != nil {
			return errMsg(err)
		}
		if s.cache != nil {
			s.cache.Set(key, items, CacheTTL)
		}
		return topicsMsg(items)
	}
}

func (s *Service) fetchSubsCmd(force bool) tea.Cmd {
	return func() tea.Msg {
		key := fmt.Sprintf("pubsub_subs:%s", s.projectID)
		if !force && s.cache != nil {
			if val, found := s.cache.Get(key); found {
				if items, ok := val.([]Subscription); ok {
					return subsMsg(items)
				}
			}
		}
		if s.client == nil {
			return errMsg(fmt.Errorf("client not init"))
		}
		items, err := s.client.ListSubscriptions(s.projectID)
		if err != nil {
			return errMsg(err)
		}
		if s.cache != nil {
			s.cache.Set(key, items, CacheTTL)
		}
		return subsMsg(items)
	}
}

// -----------------------------------------------------------------------------
// Table Updates
// -----------------------------------------------------------------------------

func (s *Service) updateTopicTable(items []Topic) {
	// Reconfigure cols for Topics
	columns := []table.Column{
		{Title: "Topic Name (Press 's' for Subs)", Width: 40},
		{Title: "KMS Key", Width: 30},
	}
	// Clear rows first to prevent panic on column resize
	s.table.SetRows([]table.Row{})
	s.table.SetColumns(columns)

	rows := make([]table.Row, len(items))
	for i, item := range items {
		rows[i] = table.Row{item.Name, item.KmsKeyName}
	}
	s.table.SetRows(rows)
}

func (s *Service) updateSubTable(items []Subscription) {
	// Reconfigure cols for Subs
	columns := []table.Column{
		{Title: "Subscription (Press 't' for Topics)", Width: 35},
		{Title: "Topic", Width: 25},
		{Title: "Type", Width: 10},
		{Title: "Ack Deadline", Width: 15},
	}
	// Clear rows first to prevent panic on column resize
	s.table.SetRows([]table.Row{})
	s.table.SetColumns(columns)

	rows := make([]table.Row, len(items))
	for i, item := range items {
		subType := "Pull"
		if item.PushEndpoint != "" {
			subType = "Push"
		}
		if item.DeadLetterTopic != "" {
			subType += " (DLQ)"
		}

		rows[i] = table.Row{
			item.Name,
			item.Topic,
			subType,
			fmt.Sprintf("%d sec", item.AckDeadline),
		}
	}
	s.table.SetRows(rows)
}
