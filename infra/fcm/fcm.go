package fcm

import (
	"context"
	"fmt"
	"shareway/schemas"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

type FCMClient struct {
	client *messaging.Client
}

func NewFCMClient(ctx context.Context, fcmConfigPath string) (*FCMClient, error) {
	opt := option.WithCredentialsFile(fcmConfigPath)
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, fmt.Errorf("error initializing app: %w", err)
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting Messaging client: %w", err)
	}

	return &FCMClient{client: client}, nil
}

func (f *FCMClient) SendNotification(ctx context.Context, notification schemas.Notification) error {
	// message := &messaging.Message{
	// 	Notification: &messaging.Notification{
	// 		Title: title,
	// 		Body:  body,
	// 	},
	// 	Token: token, // Device token from the client
	// }
	var message *messaging.Message
	if notification.Data != nil {
		message = &messaging.Message{
			Notification: &messaging.Notification{
				Title: notification.Title,
				Body:  notification.Body,
			},
			Data:  notification.Data,
			Token: notification.Token,
		}
	} else {
		message = &messaging.Message{
			Notification: &messaging.Notification{
				Title: notification.Title,
				Body:  notification.Body,
			},
			Token: notification.Token,
		}
	}

	_, err := f.client.Send(ctx, message)
	if err != nil {
		return fmt.Errorf("error sending message: %w", err)
	}

	return nil
}
