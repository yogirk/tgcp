package pubsub

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/api/pubsub/v1"
)

type Client struct {
	service *pubsub.Service
}

func NewClient(ctx context.Context) (*Client, error) {
	svc, err := pubsub.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("pubsub client: %w", err)
	}
	return &Client{service: svc}, nil
}

func (c *Client) ListTopics(projectID string) ([]Topic, error) {
	var topics []Topic
	parent := fmt.Sprintf("projects/%s", projectID)

	err := c.service.Projects.Topics.List(parent).Pages(context.Background(), func(page *pubsub.ListTopicsResponse) error {
		for _, t := range page.Topics {
			topics = append(topics, Topic{
				Name:       shortName(t.Name),
				ProjectID:  projectID,
				Labels:     t.Labels,
				KmsKeyName: t.KmsKeyName,
			})
		}
		return nil
	})
	return topics, err
}

func (c *Client) ListSubscriptions(projectID string) ([]Subscription, error) {
	var subs []Subscription
	parent := fmt.Sprintf("projects/%s", projectID)

	err := c.service.Projects.Subscriptions.List(parent).Pages(context.Background(), func(page *pubsub.ListSubscriptionsResponse) error {
		for _, s := range page.Subscriptions {
			dlTopic := ""
			if s.DeadLetterPolicy != nil {
				dlTopic = shortName(s.DeadLetterPolicy.DeadLetterTopic)
			}

			pushEp := ""
			if s.PushConfig != nil {
				pushEp = s.PushConfig.PushEndpoint
			}

			subs = append(subs, Subscription{
				Name:              shortName(s.Name),
				Topic:             shortName(s.Topic),
				PushEndpoint:      pushEp,
				AckDeadline:       int(s.AckDeadlineSeconds),
				RetainAcked:       s.RetainAckedMessages,
				RetentionDuration: s.MessageRetentionDuration,
				DeadLetterTopic:   dlTopic,
				State:             s.State,
			})
		}
		return nil
	})
	return subs, err
}

func shortName(longName string) string {
	parts := strings.Split(longName, "/")
	return parts[len(parts)-1]
}
