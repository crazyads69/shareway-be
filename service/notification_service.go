package service

import (
	"context"
	"log"
	"shareway/infra/rabbitmq"
	"shareway/repository"
	"shareway/schemas"
	"shareway/util"

	"github.com/google/uuid"
)

type INotificationService interface {
	CreateNotification(ctx context.Context, req schemas.CreateNotificationRequest, userID uuid.UUID) (uuid.UUID, error)
}

type NotificationService struct {
	repo     repository.INotificationRepository
	cfg      util.Config
	rabbitmq *rabbitmq.RabbitMQ
}

func NewNotificationService(repo repository.INotificationRepository, cfg util.Config, rabbitmq *rabbitmq.RabbitMQ) INotificationService {
	return &NotificationService{
		repo:     repo,
		cfg:      cfg,
		rabbitmq: rabbitmq,
	}
}

func (ns *NotificationService) CreateNotification(ctx context.Context, req schemas.CreateNotificationRequest, userID uuid.UUID) (uuid.UUID, error) {
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

	notification := schemas.Notification{
		Data:  req.Data,
		Title: req.Title,
		Body:  req.Body,
		Token: deviceToken,
	}

	// Publish the notification to the RabbitMQ exchange
	if err := ns.rabbitmq.PublishNotification(ctx, notification); err != nil {
		// Log the error not to interrupt the flow of the application
		// This is an asynchronous operation
		log.Printf("Failed to publish notification to RabbitMQ: %v", err)
	}

	return notificationID, nil
}
