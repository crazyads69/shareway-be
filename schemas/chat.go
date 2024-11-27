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
	MessageID   uuid.UUID `json:"message_id"`
	Message     string    `json:"message"`
	ReceiverID  uuid.UUID `json:"receiver_id"`
	SenderID    uuid.UUID `json:"sender_id"`
	MessageType string    `json:"message_type"` // text or image or call and missed_call
	CreatedAt   time.Time `json:"created_at"`
}

// Define SendImageRequest schema
type SendImageRequest struct {
	ChatRoomID string                `form:"chatRoomID" binding:"required,uuid" validate:"required,uuid"`
	Image      *multipart.FileHeader `form:"image" binding:"required" validate:"required"`
	ReceiverID string                `form:"receiverID" binding:"required,uuid" validate:"required,uuid"`
}

type SendImageResponse struct {
	ReceiverID  uuid.UUID `json:"receiver_id"`
	SenderID    uuid.UUID `json:"sender_id"`
	CreatedAt   time.Time `json:"created_at"`
	MessageType string    `json:"message_type"` // text or image or call and missed_call
	MessageID   uuid.UUID `json:"message_id"`
	Message     string    `json:"message"`
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
	MessageType string    `json:"message_type"` // text or image or call and missed_call
}

// Define InitiateCallRequest schema as query parameters
type InitiateCallRequest struct {
	ChatRoomID string `form:"chatRoomID" binding:"required,uuid"`
	ReceiverID string `form:"receiverID" binding:"required,uuid"`
}

// Define InitiateCallResponse schema
type InitiateCallResponse struct {
	Token      string    `json:"token"`
	ChatRoomID uuid.UUID `json:"chatroom_id"`
	CallID     uuid.UUID `json:"call_id"` // the call id is the message id of the chat message
	CallerID   uuid.UUID `json:"caller_id"`
	ReceiverID uuid.UUID `json:"receiver_id"`
	// CallType        string    `json:"call_type"`
}

// Define UpdateCallStatusRequest schema
type UpdateCallStatusRequest struct {
	ChatRoomID uuid.UUID `json:"chatRoomID" binding:"required,uuid" validate:"required,uuid"`
	CallType   string    `json:"callType" binding:"required,oneof=call missed_call" validate:"required,oneof=call missed_call"`
	Duration   int64     `json:"duration" validate:"min=0"` // Remove required validation because it is not parse from the validation if it is 0
	ReceiverID uuid.UUID `json:"receiverID" binding:"required,uuid" validate:"required,uuid"`
	CallID     uuid.UUID `json:"callID" binding:"required,uuid" validate:"required,uuid"`
}

// Define UpdateCallStatusResponse schema
type UpdateCallStatusResponse struct {
	ChatRoomID  uuid.UUID `json:"chat_room_id"`
	Message     string    `json:"message"`   // Call status message
	SenderID    uuid.UUID `json:"sender_id"` // User who initiated the call
	MessageID   uuid.UUID `json:"message_id"`
	CreatedAt   time.Time `json:"created_at"`   // Call status update time
	ReceiverID  uuid.UUID `json:"receiver_id"`  // User who received the call
	MessageType string    `json:"message_type"` // text or image or missed_call or call
}
