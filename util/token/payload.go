package token

import (
	"errors"
	"fmt"
	"shareway/schemas"
	"time"

	"github.com/google/uuid"
)

var (
	ErrExpiredToken = errors.New("token has expired")
)

// NewPayload creates a new token payload
// It takes a phone number and duration as input and returns a Payload and an error
func NewPayload(phoneNumber string, userID uuid.UUID, duration time.Duration) (*schemas.Payload, error) {
	// Generate a new random UUID for the token
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, fmt.Errorf("failed to generate token ID: %w", err)
	}

	// Create the current timestamp
	now := time.Now().UTC()

	// Create and return the new payload
	return &schemas.Payload{
		ID:          tokenID,
		PhoneNumber: phoneNumber,
		UserID:      userID,
		CreatedAt:   now,
		ExpiredAt:   now.Add(duration),
	}, nil
}

// ValidatePayload checks if the payload is still valid
// It returns an error if the token has expired
func ValidatePayload(payload *schemas.Payload) error {
	if payload.ExpiredAt.Before(time.Now().UTC()) {
		return ErrExpiredToken
	}
	return nil
}
