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
