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
