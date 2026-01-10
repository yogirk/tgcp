package pubsub

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/yogirk/tgcp/internal/styles"
	"github.com/yogirk/tgcp/internal/ui/components"
)

func (s *Service) View() string {
	if s.loading {
		return components.RenderSpinner("Loading Pub/Sub...")
	}
	if s.err != nil {
		return components.RenderError(s.err, "Pub/Sub", "Topics")
	}

	if s.viewState == ViewDetailTopic {
		return s.renderDetailTopic()
	}
	if s.viewState == ViewDetailSub {
		return s.renderDetailSub()
	}

	// Filter Bar
	var content strings.Builder
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

	header := lipgloss.JoinHorizontal(lipgloss.Left,
		styles.BaseStyle.Foreground(styles.ColorPrimary).Render("📡 "),
		styles.HeaderStyle.Render(fmt.Sprintf("Topic: %s", t.Name)),
	)

	details := fmt.Sprintf("Project: %s\nKMS Key: %s", t.ProjectID, t.KmsKeyName)
	box := styles.BoxStyle.Copy().Width(60).Render(
		lipgloss.JoinVertical(lipgloss.Left, header, " ", details),
	)
	return box
}

func (s *Service) renderDetailSub() string {
	sub := s.selectedSub
	if sub == nil {
		return ""
	}

	statusColor := styles.ColorSuccess
	if sub.DeadLetterTopic != "" {
		statusColor = styles.ColorWarning // Warn if DLQ configured (just for visibility)
	}

	header := lipgloss.JoinHorizontal(lipgloss.Left,
		styles.BaseStyle.Foreground(statusColor).Render("📨 "),
		styles.HeaderStyle.Render(fmt.Sprintf("Subscription: %s", sub.Name)),
	)

	dlqMsg := "None"
	if sub.DeadLetterTopic != "" {
		dlqMsg = styles.ErrorStyle.Render(sub.DeadLetterTopic)
	}

	details := fmt.Sprintf(
		"Topic: %s\nType: %s\nAck Deadline: %d sec\nRetain Acked: %v\nDead Letter Topic: %s",
		sub.Topic,
		func() string {
			if sub.PushEndpoint != "" {
				return "Push (" + sub.PushEndpoint + ")"
			} else {
				return "Pull"
			}
		}(),
		sub.AckDeadline,
		sub.RetainAcked,
		dlqMsg,
	)

	box := styles.BoxStyle.Copy().Width(80).Render(
		lipgloss.JoinVertical(lipgloss.Left, header, " ", details),
	)
	return box
}
