package schemas

import (
	"mime/multipart"
	"time"

	"github.com/google/uuid"
)

// Define SendMessageRequest schema
type SendMessageRequest struct {
	ChatRoomID uuid.UUID `json:"chatRoomID" binding:"required"`
	Message    string    `json:"message" binding:"required"` // Default type is "text"
	ReceiverID uuid.UUID `json:"receiverID" binding:"required"`
}

// Define SendMessageResponse schema
type SendMessageResponse struct {
	MessageID  uuid.UUID `json:"messageID"`
	Message    string    `json:"message"`
	ReceiverID uuid.UUID `json:"receiverID"`
	CreatedAt  time.Time `json:"createdAt"`
}

// Define SendImageRequest schema
type SendImageRequest struct {
	ChatRoomID string                `form:"chatRoomID" binding:"required,uuid" validate:"required,uuid"`
	Image      *multipart.FileHeader `form:"image" binding:"required" validate:"required"`
	ReceiverID string                `form:"receiverID" binding:"required,uuid" validate:"required,uuid"`
}

type SendImageResponse struct {
	ImageURL   string    `json:"image_url"`
	ReceiverID string    `json:"receiverID"`
	CreatedAt  time.Time `json:"createdAt"`
	MessageID  string    `json:"messageID"`
}

// Define GetAllChatRoomsRequest schema
type GetAllChatRoomsRequest struct {
	UserID uuid.UUID `json:"userID" binding:"required"`
}

// Define GetAllChatRoomsResponse schema
type GetAllChatRoomsResponse struct {
	ChatRooms []ChatRoomResponse `json:"chatRooms"`
}

// Define ChatRoomResponse schema
type ChatRoomResponse struct {
	ID            uuid.UUID `json:"room_id"`
	ReceiverInfo  UserInfo  `json:"receiver_info"`
	LastMessage   string    `json:"last_message"`
	LastMessageAt time.Time `json:"last_message_at"`
	LastMessageID uuid.UUID `json:"last_message_id"`
}

// Define GetChatMessagesRequest schema
type GetChatMessagesRequest struct {
	ChatRoomID uuid.UUID `json:"chatRoomID" binding:"required"`
}

// Define GetChatMessagesResponse schema
type GetChatMessagesResponse struct {
	Messages []MessageResponse `json:"messages"`
}

// Define MessageResponse schema

type MessageResponse struct {
	ID          uuid.UUID `json:"message_id"`
	Message     string    `json:"message"`
	SenderID    uuid.UUID `json:"sender_id"`
	CreatedAt   time.Time `json:"created_at"`
	ReceiverID  uuid.UUID `json:"receiver_id"`
	MessageType string    `json:"message_type"` // text or image or missed_call, video_call, voice_call
}
