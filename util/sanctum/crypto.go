package sanctum

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"shareway/util"
	"time"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/exp/rand"
)

type ICryptoSanctum interface {
	SHA256(plainText string) string
	HashPassword(password string) (string, error)
	VerifyPassword(hashedPassword string, password string) bool
	SercureCompare(a, b string) bool // Securely compare two strings (constant time compare for password)
}

type CryptoSanctum struct {
	cfg util.Config
}

func NewCryptoSanctum(
	cfg util.Config,
) ICryptoSanctum {
	return &CryptoSanctum{
		cfg: cfg,
	}
}

// SHA256 is a function to hash a plain text using SHA256 algorithm
func (c *CryptoSanctum) SHA256(plainText string) string {
	hash := sha256.Sum256([]byte(plainText))
	return hex.EncodeToString(hash[:])
}

// HashPassword is a function to hash a password using SHA256 algorithm
func (c *CryptoSanctum) HashPassword(password string) (string, error) {
	if len(password) == 0 {
		return "", errors.New("password cannot be empty")
	}
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), c.cfg.BcryptCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(bytes), nil
}

// VerifyPassword is a function to verify a hashed password
func (c *CryptoSanctum) VerifyPassword(hashedPassword string, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// SecureCompare is a function to securely compare two strings (constant time compare for password) to prevent timing attacks
func (c *CryptoSanctum) SercureCompare(a, b string) bool {
	// Use subtle.ConstantTimeCompare to prevent timing attacks
	if subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1 {
		// Randomize the delay slightly to prevent timing attacks
		time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
		return true
	}
	// Randomize the delay slightly to prevent timing attacks
	time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
	return false
}

var _ ICryptoSanctum = (*CryptoSanctum)(nil)
