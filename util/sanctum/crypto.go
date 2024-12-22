package sanctum

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"shareway/util"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type ICryptoSanctum interface {
	SHA256(plainText string) string
	HMACSHA256(plainText string) string
	HashPassword(password string) (string, error)
	VerifyPassword(hashedPassword string, password string) bool
	SecureCompare(hashedPassword, password string) bool // Securely compare two strings (constant time compare for password)
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

// HMACSHA256 is a function to create an HMAC using SHA256
func (c *CryptoSanctum) HMACSHA256(plainText string) string {
	// Load secret key from environment variable and decode base64
	sanctumSecretKey, err := base64.StdEncoding.DecodeString(c.cfg.SanctumSecretKey)
	if err != nil {
		return ""
	}

	h := hmac.New(sha256.New, []byte(sanctumSecretKey))
	h.Write([]byte(plainText))
	return hex.EncodeToString(h.Sum(nil))
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
func (c *CryptoSanctum) SecureCompare(hashedPassword, password string) bool {
	// First, add a small constant delay to mask the length check
	time.Sleep(1 * time.Millisecond)

	// Convert strings to byte slices only once
	aBytes := []byte(hashedPassword)
	bBytes := []byte(password)

	// Use subtle.ConstantTimeCompare to prevent timing attacks
	return subtle.ConstantTimeCompare(aBytes, bBytes) == 1
}

var _ ICryptoSanctum = (*CryptoSanctum)(nil)
