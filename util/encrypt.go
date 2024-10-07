package util

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"sync"
)

// IEncryptor defines the interface for encryption and decryption operations
type IEncryptor interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(ciphertext string) (string, error)
}

// Encryptor implements the IEncryptor interface
type Encryptor struct {
	cfg   Config
	block cipher.Block
	once  sync.Once
}

// NewEncryptor creates a new Encryptor instance
func NewEncryptor(cfg Config) IEncryptor {
	return &Encryptor{
		cfg: cfg,
	}
}

// initBlock initializes the AES block cipher (lazy initialization)
func (e *Encryptor) initBlock() error {
	var err error
	e.once.Do(func() {
		// Decode the base64-encoded key
		key, decodeErr := base64.StdEncoding.DecodeString(e.cfg.EncryptionKey)
		if decodeErr != nil {
			err = errors.New("failed to decode encryption key")
			return
		}

		// Create the AES cipher using the decoded key
		e.block, err = aes.NewCipher(key)
	})
	return err
}

// Encrypt encrypts the given plaintext
func (e *Encryptor) Encrypt(plaintext string) (string, error) {
	if err := e.initBlock(); err != nil {
		return "", err
	}

	// Create a byte slice with room for the IV and plaintext
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]

	// Fill the IV with random bytes
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	// Encrypt the plaintext
	stream := cipher.NewCFBEncrypter(e.block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], []byte(plaintext))

	// Encode the result as base64
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts the given ciphertext
func (e *Encryptor) Decrypt(encryptedText string) (string, error) {
	if err := e.initBlock(); err != nil {
		return "", err
	}

	// Decode the base64 encoded ciphertext
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedText)
	if err != nil {
		return "", err
	}

	// Ensure the ciphertext is long enough to contain the IV
	if len(ciphertext) < aes.BlockSize {
		return "", errors.New("ciphertext too short")
	}

	// Extract the IV and the actual ciphertext
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	// Decrypt the ciphertext
	stream := cipher.NewCFBDecrypter(e.block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	return string(ciphertext), nil
}

// Ensure Encryptor implements IEncryptor
var _ IEncryptor = (*Encryptor)(nil)
