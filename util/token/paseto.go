package token

import (
	"encoding/base64"
	"fmt"
	"time"

	"shareway/schemas"

	"github.com/aead/chacha20poly1305"
	"github.com/google/uuid"
	"github.com/o1egl/paseto"
)

// PasetoMaker is responsible for creating and verifying PASETO tokens
type PasetoMaker struct {
	paseto       *paseto.V2
	symmetricKey []byte
}

func SetupPasetoMaker(base64Key string) (*PasetoMaker, error) {
	key, err := base64.StdEncoding.DecodeString(base64Key)
	if err != nil {
		return nil, fmt.Errorf("invalid key: %w", err)
	}
	return NewPasetoMaker(string(key))
}

// NewPasetoMaker creates a new PasetoMaker instance
// It takes a symmetricKey as input and returns a PasetoMaker and an error
func NewPasetoMaker(symmetricKey string) (*PasetoMaker, error) {
	// Check if the symmetric key has the correct length
	if len(symmetricKey) != chacha20poly1305.KeySize {
		return nil, fmt.Errorf("invalid key size: expected %d, got %d", chacha20poly1305.KeySize, len(symmetricKey))
	}

	// Create and return a new PasetoMaker
	return &PasetoMaker{
		paseto:       paseto.NewV2(),
		symmetricKey: []byte(symmetricKey),
	}, nil
}

// CreateToken generates a new token for a given phone number and duration
func (maker *PasetoMaker) CreateToken(phoneNumber string, userID uuid.UUID, duration time.Duration) (string, error) {
	// Create a new payload
	payload, err := NewPayload(phoneNumber, userID, duration)
	if err != nil {
		return "", fmt.Errorf("failed to create payload: %w", err)
	}

	// Encrypt the payload and return the token
	return maker.paseto.Encrypt(maker.symmetricKey, payload, nil)
}

// VerifyToken checks if a token is valid and returns the payload
func (maker *PasetoMaker) VerifyToken(token string) (*schemas.Payload, error) {
	var payload schemas.Payload

	// Decrypt the token
	if err := maker.paseto.Decrypt(token, maker.symmetricKey, &payload, nil); err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	// Validate the payload
	if err := ValidatePayload(&payload); err != nil {
		return nil, fmt.Errorf("invalid payload: %w", err)
	}

	return &payload, nil
}
