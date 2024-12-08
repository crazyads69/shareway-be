package service

import (
	"context"
	"shareway/infra/bucket"
	"shareway/infra/db/migration"
	"shareway/infra/ws"
	"shareway/repository"
	"shareway/schemas"
	"shareway/util"

	"github.com/google/uuid"
)

type ChatService struct {
	repo       repository.IChatRepository
	hub        *ws.Hub
	cfg        util.Config
	cloudinary *bucket.CloudinaryService
}

func NewChatService(repo repository.IChatRepository, hub *ws.Hub, cfg util.Config, cloudinary *bucket.CloudinaryService) IChatService {
	return &ChatService{
		repo:       repo,
		hub:        hub,
		cfg:        cfg,
		cloudinary: cloudinary,
	}
}

type IChatService interface {
	SendMessage(req schemas.SendMessageRequest, userID uuid.UUID) (migration.Chat, error)
	UploadImage(ctx context.Context, req schemas.SendImageRequest, userID uuid.UUID) (migration.Chat, error)
	GetAllChatRooms(userID uuid.UUID) ([]migration.Room, error)
	GetChatMessages(req schemas.GetChatMessagesRequest, userID uuid.UUID) ([]migration.Chat, error)
	UpdateCallStatus(req schemas.UpdateCallStatusRequest, userID uuid.UUID) (migration.Chat, error)
	InitiateCall(req schemas.InitiateCallRequest, userID uuid.UUID) (migration.Chat, error)
	SearchUsers(req schemas.SearchUsersRequest, userID uuid.UUID) ([]migration.Room, error)
}

// SendMessage sends a message to a chat room
func (s *ChatService) SendMessage(req schemas.SendMessageRequest, userID uuid.UUID) (migration.Chat, error) {
	return s.repo.SendMessage(req, userID)
}

// UploadImage uploads an image to a chat room
func (s *ChatService) UploadImage(ctx context.Context, req schemas.SendImageRequest, userID uuid.UUID) (migration.Chat, error) {
	// First, upload the image to Cloudinary
	imageURL, err := s.cloudinary.UploadChatImage(ctx, req.Image)
	if err != nil {
		return migration.Chat{}, err
	}

	// Then, send the message to the chat room
	chat, err := s.repo.UploadImage(req, userID, imageURL)
	if err != nil {
		return migration.Chat{}, err
	}

	return chat, nil
}

// GetAllChatRooms fetches all chat rooms for a user
func (s *ChatService) GetAllChatRooms(userID uuid.UUID) ([]migration.Room, error) {
	return s.repo.GetAllChatRooms(userID)
}

// GetChatMessages fetches all messages in a chat room
func (s *ChatService) GetChatMessages(req schemas.GetChatMessagesRequest, userID uuid.UUID) ([]migration.Chat, error) {
	return s.repo.GetChatMessages(req, userID)
}

// UpdateCallStatus updates the call status in a chat room
func (s *ChatService) UpdateCallStatus(req schemas.UpdateCallStatusRequest, userID uuid.UUID) (migration.Chat, error) {
	return s.repo.UpdateCallStatus(req, userID)
}

// InitiateCall initiates a call in a chat room
func (s *ChatService) InitiateCall(req schemas.InitiateCallRequest, userID uuid.UUID) (migration.Chat, error) {
	return s.repo.InitiateCall(req, userID)
}

// SearchUsers searches for users to chat with
func (s *ChatService) SearchUsers(req schemas.SearchUsersRequest, userID uuid.UUID) ([]migration.Room, error) {
	return s.repo.SearchUsers(req, userID)
}

// Ensure ChatService implements IChatService
var _ IChatService = (*ChatService)(nil)
