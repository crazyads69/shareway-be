package repository

import (
	"shareway/infra/db/migration"
	"shareway/schemas"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type ChatRepository struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewChatRepository(db *gorm.DB, redis *redis.Client) IChatRepository {
	return &ChatRepository{
		db:    db,
		redis: redis,
	}
}

type IChatRepository interface {
	SendMessage(req schemas.SendMessageRequest, userID uuid.UUID) (migration.Chat, error)
	GetChatRoomByUserIDs(userID1, userID2 uuid.UUID) (migration.Room, error)
	UploadImage(req schemas.SendImageRequest, userID uuid.UUID, imageURL string) (migration.Chat, error)
	GetAllChatRooms(userID uuid.UUID) ([]migration.Room, error)
	GetChatMessages(req schemas.GetChatMessagesRequest, userID uuid.UUID) ([]migration.Chat, error)
}

// GetChatRoomByUserIDs fetches a chat room by the user IDs
func (r *ChatRepository) GetChatRoomByUserIDs(userID1, userID2 uuid.UUID) (migration.Room, error) {
	var room migration.Room

	// Ensure user IDs are in a consistent order
	if userID1.String() > userID2.String() {
		userID1, userID2 = userID2, userID1
	}

	// Check if the chat room already exists
	err := r.db.Model(&migration.Room{}).
		Where("user1_id = ? AND user2_id = ?", userID1, userID2).
		Or("user1_id = ? AND user2_id = ?", userID2, userID1).
		First(&room).Error

	if err != nil {
		return migration.Room{}, err
	}

	return room, err
}

// SendMessage sends a message to a chat room
func (r *ChatRepository) SendMessage(req schemas.SendMessageRequest, userID uuid.UUID) (migration.Chat, error) {

	// Retrieve the chat room
	chat, err := r.GetChatRoomByUserIDs(userID, req.ReceiverID)
	if err != nil {
		return migration.Chat{}, err
	}

	// Create a new chat message
	newChat := migration.Chat{
		RoomID:      chat.ID,
		SenderID:    userID,
		ReceiverID:  req.ReceiverID,
		Message:     req.Message,
		MessageType: "text",
	}

	// Save the chat message
	err = r.db.Create(&newChat).Error
	if err != nil {
		return migration.Chat{}, err
	}

	// Update the chat room lastest message
	// Update the chat room with the last message ID and message content
	if err := r.db.Model(&migration.Room{}).Where("id = ?", chat.ID).Updates(map[string]interface{}{
		"last_message_id":   newChat.ID,
		"last_message_text": newChat.Message,
		"last_message_at":   newChat.CreatedAt,
	}).Error; err != nil {
		return migration.Chat{}, err
	}

	return newChat, nil
}

// UploadImage uploads an image to a chat room
func (r *ChatRepository) UploadImage(req schemas.SendImageRequest, userID uuid.UUID, imageURL string) (migration.Chat, error) {
	// Parse the receiver ID
	receiverID, err := uuid.Parse(req.ReceiverID)
	if err != nil {
		return migration.Chat{}, err
	}
	// Retrieve the chat room
	chat, err := r.GetChatRoomByUserIDs(userID, receiverID)
	if err != nil {
		return migration.Chat{}, err
	}

	// Create a new chat message
	newChat := migration.Chat{
		RoomID:      chat.ID,
		SenderID:    userID,
		ReceiverID:  receiverID,
		Message:     imageURL,
		MessageType: "image",
	}

	// Save the chat message
	err = r.db.Create(&newChat).Error
	if err != nil {
		return migration.Chat{}, err
	}

	// Update the chat room lastest message
	// Update the chat room with the last message ID and message content
	if err := r.db.Model(&migration.Room{}).Where("id = ?", chat.ID).Updates(map[string]interface{}{
		"last_message_id":   newChat.ID,
		"last_message_text": "Hình ảnh", // Display "Hình ảnh" for image messages not to show the image URL
		"last_message_at":   newChat.CreatedAt,
	}).Error; err != nil {
		return migration.Chat{}, err
	}

	return newChat, nil
}

// GetAllChatRooms fetches all chat rooms for a user
func (r *ChatRepository) GetAllChatRooms(userID uuid.UUID) ([]migration.Room, error) {
	var rooms []migration.Room

	// Fetch all chat rooms for the user
	err := r.db.Model(&migration.Room{}).
		Where("user1_id = ?", userID).
		Or("user2_id = ?", userID).
		Find(&rooms).Error

	if err != nil {
		return nil, err
	}

	return rooms, nil
}

// GetChatMessages fetches all messages in a chat room
func (r *ChatRepository) GetChatMessages(req schemas.GetChatMessagesRequest, userID uuid.UUID) ([]migration.Chat, error) {
	var messages []migration.Chat

	// Fetch all messages in the chat room
	err := r.db.Model(&migration.Chat{}).
		Where("room_id = ?", req.ChatRoomID).
		Order("created_at ASC"). // Order by created_at in ascending order to get older messages first
		Find(&messages).Error

	if err != nil {
		return nil, err
	}

	return messages, nil
}
