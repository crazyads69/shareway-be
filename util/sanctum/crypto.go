package sanctum

import (
	"crypto/sha256"
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"
)

type ICryptoSanctum interface {
	SHA256(plainText string) string
	HashPassword(password string) string
	VerifyPassword(hashedPassword string, password string) bool
}

type CryptoSanctum struct {
}

func NewCryptoSanctum() ICryptoSanctum {
	return &CryptoSanctum{}
}

// SHA256 is a function to hash a plain text using SHA256 algorithm
func (c *CryptoSanctum) SHA256(plainText string) string {
	hash := sha256.Sum256([]byte(plainText))
	return hex.EncodeToString(hash[:])
}

// HashPassword is a function to hash a password using SHA256 algorithm
func (c *CryptoSanctum) HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return ""
	}
	return string(bytes)
}

// VerifyPassword is a function to verify a hashed password
func (c *CryptoSanctum) VerifyPassword(hashedPassword string, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

var _ ICryptoSanctum = (*CryptoSanctum)(nil)
