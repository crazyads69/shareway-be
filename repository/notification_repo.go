package repository

import (
	"shareway/infra/db/migration"
	"shareway/schemas"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type INotificationRepository interface {
	CreateNotification(req schemas.CreateNotificationRequest, userID uuid.UUID) (uuid.UUID, error)
	GetUserDeviceToken(userID uuid.UUID) (string, error)
}

type NotificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) INotificationRepository {
	return &NotificationRepository{
		db: db,
	}
}

func (nr *NotificationRepository) CreateNotification(req schemas.CreateNotificationRequest, userID uuid.UUID) (uuid.UUID, error) {
	// Convert Data to JSONB type (from map[string]string to map[string]interface{})
	data := make(map[string]interface{})
	for k, v := range req.Data {
		data[k] = v
	}

	token, err := nr.GetUserDeviceToken(userID)
	if err != nil {
		return uuid.Nil, err
	}

	notification := migration.Notification{
		UserID:   userID,
		Title:    req.Title,
		Body:     req.Body,
		Data:     data,
		TokenFCM: token,
	}

	if err := nr.db.Create(&notification).Error; err != nil {
		return uuid.Nil, err
	}

	return notification.ID, nil
}

func (nr *NotificationRepository) GetUserDeviceToken(userID uuid.UUID) (string, error) {
	var user migration.User
	if err := nr.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return "", err
	}

	return user.DeviceToken, nil
}
