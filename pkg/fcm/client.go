package fcm

import (
	"context"
	"log"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

type Notification struct {
	Title string
	Body  string
}

type Message struct {
	Token        string
	Notification *Notification
	Data         map[string]string
}

type Client struct {
	msg *messaging.Client
}

// NewClient initialises the Firebase Admin SDK.
// If the credentials file is missing, the client degrades gracefully
// (notifications are skipped instead of crashing).
func NewClient(credentialsPath string) *Client {
	ctx := context.Background()
	app, err := firebase.NewApp(ctx, nil, option.WithCredentialsFile(credentialsPath))
	if err != nil {
		log.Printf("fcm: failed to init app (%v) — push notifications disabled", err)
		return &Client{}
	}
	msgClient, err := app.Messaging(ctx)
	if err != nil {
		log.Printf("fcm: failed to get messaging client (%v) — push notifications disabled", err)
		return &Client{}
	}
	return &Client{msg: msgClient}
}

func (c *Client) Send(ctx context.Context, m *Message) error {
	if c.msg == nil {
		log.Printf("fcm: skipping notification to %s (client not initialised)", m.Token)
		return nil
	}

	fcmMsg := &messaging.Message{
		Token: m.Token,
		Data:  m.Data,
	}
	if m.Notification != nil {
		fcmMsg.Notification = &messaging.Notification{
			Title: m.Notification.Title,
			Body:  m.Notification.Body,
		}
	}

	_, err := c.msg.Send(ctx, fcmMsg)
	return err
}
