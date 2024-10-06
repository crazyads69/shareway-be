package schemas

import (
	"time"

	"github.com/google/uuid"
)

// Define Payload struct

type Payload struct {
	ID          uuid.UUID `json:"id"`
	PhoneNumber string    `json:"phone_number"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiredAt   time.Time `json:"expired_at"`
}

// Define GenerateOTPRequest struct
type GenerateOTPRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required,numeric,min=9,max=11"`
}

// Define GenerateOTPResponse struct
type GenerateOTPResponse struct {
	OTP string `json:"otp" binding:"required,numeric,min=6,max=6"`
}

// Define struct to first register a user
type FirstRegisterUserRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required,numeric,min=9,max=11"`
}

// Define struct to first register a user response
type FirstRegisterUserResponse struct {
	UserID      uuid.UUID `json:"user_id" binding:"required"`
	PhoneNumber string    `json:"phone_number" binding:"required,numeric,min=9,max=11"`
}

// Define struct to verify the OTP
type VerifyRegisterTPRequest struct {
	PhoneNumber string    `json:"phone_number" binding:"required,numeric,min=9,max=11"`
	UserID      uuid.UUID `json:"user_id" binding:"required"`
}

// Define struct to verify the OTP response
type VerifyRegisterOTPResponse struct {
	UserID      uuid.UUID `json:"user_id" binding:"required"`
	PhoneNumber string    `json:"phone_number" binding:"required,numeric,min=9,max=11"`
	IsVerified  bool      `json:"is_verified" binding:"required"`
}

// Define struct to login a user
type LoginUserRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required,numeric,min=9,max=11"`
}

// Define struct to login a user response
type LoginUserResponse struct {
	UserID      uuid.UUID `json:"user_id" binding:"required"`
	PhoneNumber string    `json:"phone_number" binding:"required,numeric,min=9,max=11"`
}

// Define struct to verify the OTP
type VerifyLoginOTPRequest struct {
	PhoneNumber string    `json:"phone_number" binding:"required,numeric,min=9,max=11"`
	UserID      uuid.UUID `json:"user_id" binding:"required"`
}
