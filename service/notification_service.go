package service

import (
	"fmt"
	"log"
	"shareway/repository"
	"shareway/schemas"
	"shareway/util"

	"github.com/google/uuid"
)

type INotificationService interface {
	CreateNotification(req schemas.CreateNotificationRequest, userID uuid.UUID) (uuid.UUID, error)
	CreateTestWebsocket(req schemas.CreateTestWebsocketRequest, userID uuid.UUID) error
}

type NotificationService struct {
	repo repository.INotificationRepository
	cfg  util.Config
}

func NewNotificationService(repo repository.INotificationRepository, cfg util.Config) INotificationService {
	return &NotificationService{
		repo: repo,
		cfg:  cfg,
	}
}

func (ns *NotificationService) CreateNotification(req schemas.CreateNotificationRequest, userID uuid.UUID) (uuid.UUID, error) {
	// Save the notification to the database
	notificationID, err := ns.repo.CreateNotification(req, userID)
	if err != nil {
		return uuid.Nil, err
	}

	// Get the device token of the user
	deviceToken, err := ns.repo.GetUserDeviceToken(userID)
	if err != nil {
		return uuid.Nil, err
	}

	// Check if the user has a device token
	if deviceToken == "" {
		return uuid.Nil, fmt.Errorf("user does not have a device token")
	}

	notification := schemas.Notification{
		Data:  req.Data,
		Title: req.Title,
		Body:  req.Body,
		Token: deviceToken,
	}

	log.Printf("Sending notification to device: %v", notification)

	// Publish the notification to the RabbitMQ exchange
	// This is a asynchronous task so much run in a goroutine
	// go func() {
	// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// 	defer cancel()

	// 	err := ns.rabbitmq.PublishNotification(ctx, notification)
	// 	if err != nil {
	// 		log.Printf("Failed to publish notification to RabbitMQ: %v", err)
	// 		// Consider implementing a retry mechanism or storing failed notifications
	// 	}
	// }()

	return notificationID, nil
}

func (ns *NotificationService) CreateTestWebsocket(req schemas.CreateTestWebsocketRequest, userID uuid.UUID) error {
	// Prepare the message to be sent to the websocket
	message := schemas.WebSocketMessage{
		UserID:  userID.String(),
		Payload: req,
		Type:    "test",
	}

	log.Printf("Sending test websocket message: %v", message)

	// Publish the message to the RabbitMQ exchange
	// This is a asynchronous task so much run in a goroutine
	// go func() {
	// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	// 	defer cancel()

	// err := ns.rabbitmq.PublishWebSocketMessage(context.Background(), message)
	// if err != nil {
	// 	log.Printf("Failed to publish websocket message to RabbitMQ: %v", err)
	// 	// Consider implementing a retry mechanism or storing failed messages
	// }
	// // }()

	return nil
}
