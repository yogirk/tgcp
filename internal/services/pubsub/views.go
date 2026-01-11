package pubsub

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/yogirk/tgcp/internal/styles"
	"github.com/yogirk/tgcp/internal/ui/components"
)

func (s *Service) View() string {
	if s.err != nil {
		return components.RenderError(s.err, "Pub/Sub", "Topics")
	}

	// Show spinner while loading
	if s.spinner.IsActive() {
		return s.spinner.View()
	}

	if s.viewState == ViewDetailTopic {
		return s.renderDetailTopic()
	}
	if s.viewState == ViewDetailSub {
		return s.renderDetailSub()
	}

	// Filter Bar
	var content strings.Builder
	listLabel := "Topics"
	if s.viewState == ViewListSubs {
		listLabel = "Subscriptions"
	}
	content.WriteString(components.Breadcrumb(
		fmt.Sprintf("Project %s", s.projectID),
		s.Name(),
		listLabel,
	))
	content.WriteString("\n")
	content.WriteString(s.filter.View())
	content.WriteString("\n")
	content.WriteString(s.table.View())
	return content.String()
}

func (s *Service) renderDetailTopic() string {
	t := s.selectedTopic
	if t == nil {
		return ""
	}

	breadcrumb := components.Breadcrumb(
		fmt.Sprintf("Project %s", s.projectID),
		s.Name(),
		"Topics",
		t.Name,
	)

	card := components.DetailCard(components.DetailCardOpts{
		Title: "Topic Details",
		Rows: []components.KeyValue{
			{Key: "Name", Value: t.Name},
			{Key: "Project", Value: t.ProjectID},
			{Key: "KMS Key", Value: t.KmsKeyName},
		},
		Width: 60,
	})
	return lipgloss.JoinVertical(lipgloss.Left, breadcrumb, "", card)
}

func (s *Service) renderDetailSub() string {
	sub := s.selectedSub
	if sub == nil {
		return ""
	}

	breadcrumb := components.Breadcrumb(
		fmt.Sprintf("Project %s", s.projectID),
		s.Name(),
		"Subscriptions",
		sub.Name,
	)

	dlqMsg := "None"
	if sub.DeadLetterTopic != "" {
		dlqMsg = styles.ErrorStyle.Render(sub.DeadLetterTopic)
	}

	subType := "Pull"
	if sub.PushEndpoint != "" {
		subType = "Push (" + sub.PushEndpoint + ")"
	}

	card := components.DetailCard(components.DetailCardOpts{
		Title: "Subscription Details",
		Rows: []components.KeyValue{
			{Key: "Name", Value: sub.Name},
			{Key: "Topic", Value: sub.Topic},
			{Key: "Type", Value: subType},
			{Key: "Ack Deadline", Value: fmt.Sprintf("%d sec", sub.AckDeadline)},
			{Key: "Retain Acked", Value: fmt.Sprintf("%v", sub.RetainAcked)},
			{Key: "Dead Letter Topic", Value: dlqMsg},
		},
	})
	return lipgloss.JoinVertical(lipgloss.Left, breadcrumb, "", card)
}
