package slack

import (
	"context"
	"time"

	"github.com/slack-go/slack"
)

type Notifier interface {
	Notify(
		ctx context.Context,
		jobName string,
		podNamespace string,
		podName string,
		isSucceeded bool,
		timestamp time.Time,
	) error
}

type Client struct {
	WebhookURL string
}

// NewClient creates slack client
func NewClient(webhookURL string) *Client {
	return &Client{webhookURL}
}

// Notify makes slack message to notify CI result
func (c *Client) Notify(
	ctx context.Context,
	jobName string,
	podNamespace string,
	podName string,
	isSucceeded bool,
	timestamp time.Time,
) error {
	text := "CI Failed"
	color := "danger"
	if isSucceeded {
		text = "CI Succeeded"
		color = "good"
	}

	message := &slack.WebhookMessage{
		Attachments: []slack.Attachment{
			slack.Attachment{
				AuthorName: "Self-hosted Runner",
				Color:      color,
				Title:      "GitHub Actions",
				Text:       text,
				Fields: []slack.AttachmentField{
					{Title: "Job", Value: jobName, Short: true},
					{Title: "TimeStamp", Value: timestamp.Format(time.RFC3339), Short: true},
					{Title: "Namespace", Value: podNamespace, Short: true},
					{Title: "Pod", Value: podName, Short: true},
				},
			},
		},
	}

	return slack.PostWebhookContext(ctx, c.WebhookURL, message)
}

// TODO: Remove fakeClient because the logic is so simple that
// this turns out not to be needed
type fakeClient struct {
	WebhookURL string
}

// NewFakeClient creates client mock for test
func NewFakeClient(webhookURL string) *fakeClient {
	return &fakeClient{webhookURL}
}

// Notify just returns nil
func (c *fakeClient) Notify(
	ctx context.Context,
	jobName string,
	podNamespace string,
	podName string,
	isSucceeded bool,
	timestamp time.Time,
) error {
	return nil
}
