package pubsub

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/yogirk/tgcp/internal/core"
	"github.com/yogirk/tgcp/internal/ui/components"
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
	table     *components.StandardTable

	filter             components.FilterModel
	topicFilterSession components.FilterSession[Topic]
	subFilterSession   components.FilterSession[Subscription]

	topics []Topic
	subs   []Subscription

	spinner components.SpinnerModel
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

	t := components.NewStandardTable(columns)

	svc := &Service{
		table:     t,
		filter:    components.NewFilterWithPlaceholder("Filter topics/subscriptions..."),
		spinner:   components.NewSpinner(),
		viewState: ViewListTopics,
		cache:     cache,
	}
	svc.topicFilterSession = components.NewFilterSession(&svc.filter, svc.getFilteredTopics, svc.updateTopicTable)
	svc.subFilterSession = components.NewFilterSession(&svc.filter, svc.getFilteredSubs, svc.updateSubTable)
	return svc
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
		s.fetchTopicsCmd(true),
		s.fetchSubsCmd(true),
	)
}

func (s *Service) Reset() {
	s.viewState = ViewListTopics
	s.selectedTopic = nil
	s.selectedSub = nil
	s.err = nil
	s.table.SetCursor(0)
	s.filter.ExitFilterMode()
	s.updateTopicTable(s.topics) // Reset cols
}

func (s *Service) IsRootView() bool {
	return s.viewState == ViewListTopics || s.viewState == ViewListSubs
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
		return s, tea.Batch(s.fetchTopicsCmd(false), s.fetchSubsCmd(false), s.tick())

	case topicsMsg:
		s.topics = msg
		if s.viewState == ViewListTopics {
			s.spinner.Stop()
			s.topicFilterSession.Apply(s.topics)
		}
		return s, func() tea.Msg { return core.LastUpdatedMsg(time.Now()) }

	case subsMsg:
		s.subs = msg
		if s.viewState == ViewListSubs {
			s.spinner.Stop()
			s.subFilterSession.Apply(s.subs)
		}
		return s, func() tea.Msg { return core.LastUpdatedMsg(time.Now()) }

	case errMsg:
		s.spinner.Stop()
		s.err = msg
		return s, nil

	case tea.WindowSizeMsg:
		s.table.HandleWindowSizeDefault(msg)

	case tea.KeyMsg:
		// Handle filter mode (only in list views)
		if s.viewState == ViewListTopics || s.viewState == ViewListSubs {
			var result components.FilterUpdateResult
			if s.viewState == ViewListTopics {
				result = s.topicFilterSession.HandleKey(msg)
			} else {
				result = s.subFilterSession.HandleKey(msg)
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

		// Root Views (Topics or Subs)
		if s.viewState == ViewListTopics || s.viewState == ViewListSubs {
			switch msg.String() {
			case "r":
				return s, s.Refresh()
			case "s": // Switch to Subs
				if s.viewState == ViewListTopics {
					s.viewState = ViewListSubs
					s.subFilterSession.Apply(s.subs) // Render existing if available
					if len(s.subs) == 0 {
						return s, tea.Batch(s.fetchSubsCmd(true), s.spinner.Start(""))
					}
				}
			case "t": // Switch to Topics
				if s.viewState == ViewListSubs {
					s.viewState = ViewListTopics
					s.topicFilterSession.Apply(s.topics)
					if len(s.topics) == 0 {
						return s, tea.Batch(s.fetchTopicsCmd(true), s.spinner.Start(""))
					}
				}
			case "enter":
				if s.viewState == ViewListTopics {
					topics := s.getFilteredTopics(s.topics, s.filter.Value())
					if idx := s.table.Cursor(); idx >= 0 && idx < len(topics) {
						s.selectedTopic = &topics[idx]
						s.viewState = ViewDetailTopic
					}
				} else {
					subs := s.getFilteredSubs(s.subs, s.filter.Value())
					if idx := s.table.Cursor(); idx >= 0 && idx < len(subs) {
						s.selectedSub = &subs[idx]
						s.viewState = ViewDetailSub
					}
				}
			}
			var updatedTable *components.StandardTable
			updatedTable, cmd = s.table.Update(msg)
			s.table = updatedTable
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

// getFilteredTopics returns filtered topics based on the query string
func (s *Service) getFilteredTopics(topics []Topic, query string) []Topic {
	if query == "" {
		return topics
	}
	return components.FilterSlice(topics, query, func(topic Topic, q string) bool {
		return components.ContainsMatch(topic.Name, topic.KmsKeyName)(q)
	})
}

// getFilteredSubs returns filtered subscriptions based on the query string
func (s *Service) getFilteredSubs(subs []Subscription, query string) []Subscription {
	if query == "" {
		return subs
	}
	return components.FilterSlice(subs, query, func(sub Subscription, q string) bool {
		return components.ContainsMatch(sub.Name, sub.Topic, sub.PushEndpoint)(q)
	})
}
