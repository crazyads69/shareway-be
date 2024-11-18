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
	MessageID    uuid.UUID `json:"message_id"`
	Message      string    `json:"message"`
	ReceiverID   uuid.UUID `json:"receiver_id"`
	CallStatus   string    `json:"call_status"`   // missed or ended or rejected
	CallDuration int64     `json:"call_duration"` // Call duration in seconds
	MessageType  string    `json:"message_type"`  // text or image or missed_call, video_call, voice_call
	CreatedAt    time.Time `json:"createdAt"`
}

// Define SendImageRequest schema
type SendImageRequest struct {
	ChatRoomID string                `form:"chatRoomID" binding:"required,uuid" validate:"required,uuid"`
	Image      *multipart.FileHeader `form:"image" binding:"required" validate:"required"`
	ReceiverID string                `form:"receiverID" binding:"required,uuid" validate:"required,uuid"`
}

type SendImageResponse struct {
	ReceiverID   uuid.UUID `json:"receiver_id"`
	CreatedAt    time.Time `json:"createdAt"`
	CallStatus   string    `json:"call_status"`   // missed or ended or rejected
	CallDuration int64     `json:"call_duration"` // Call duration in seconds
	MessageType  string    `json:"message_type"`  // text or image or missed_call, video_call, voice_call
	MessageID    uuid.UUID `json:"messageID"`
	Message      string    `json:"message"`
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
	ID           uuid.UUID `json:"message_id"`
	Message      string    `json:"message"`
	CreatedAt    time.Time `json:"created_at"`
	ReceiverID   uuid.UUID `json:"receiver_id"`
	CallStatus   string    `json:"call_status"`   // missed or ended or rejected
	CallDuration int64     `json:"call_duration"` // Call duration in seconds
	MessageType  string    `json:"message_type"`  // text or image or missed_call, video_call, voice_call
}

// Define InitiateCallRequest schema as query parameters
type InitiateCallRequest struct {
	ChatRoomID uuid.UUID `json:"chatRoomID" binding:"required,uuid" validate:"required,uuid"`
	Role       string    `json:"role" binding:"required" validate:"required,oneof=publisher subscriber"` // publisher or subscriber
	ReceiverID uuid.UUID `json:"receiverID" binding:"required" validate:"required,uuid"`
	ExpireTime uint32    `json:"expireTime" binding:"required" validate:"required"`
	// CallType   string    `json:"callType" binding:"required" validate:"required,oneof=video_call voice_call"`
}

// Define InitiateCallResponse schema
type InitiateCallResponse struct {
	TokenPublisher  string    `json:"token_publisher"`
	TokenSubscriber string    `json:"token_subscriber"`
	ChatRoomID      uuid.UUID `json:"chatroom_id"`
	CallerID        uuid.UUID `json:"caller_id"`
	// CallType        string    `json:"call_type"`
}

// Define UpdateCallStatusRequest schema
type UpdateCallStatusRequest struct {
	ChatRoomID uuid.UUID `json:"chatRoomID" binding:"required"`
	CallStatus string    `json:"callStatus" binding:"required" validate:"required,oneof=missed ended rejected"`
	CallType   string    `json:"callType" binding:"required" validate:"required,oneof=video_call voice_call"`
	Duration   int64     `json:"duration" binding:"omitempty"`
	ReceiverID uuid.UUID `json:"receiverID" binding:"required"`
}

// Define UpdateCallStatusResponse schema
type UpdateCallStatusResponse struct {
	ChatRoomID   uuid.UUID `json:"chat_room_id"`
	CallStatus   string    `json:"call_status"`   // missed or ended or rejected
	CallDuration int64     `json:"call_duration"` // Call duration in seconds
	Message      string    `json:"message"`       // Call status message
	MessageID    uuid.UUID `json:"message_id"`
	CreatedAt    time.Time `json:"created_at"`   // Call status update time
	ReceiverID   uuid.UUID `json:"receiver_id"`  // User who received the call
	MessageType  string    `json:"message_type"` // text or image or missed_call, video_call, voice_call
}
