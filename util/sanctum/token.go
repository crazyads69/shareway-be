package sanctum

import (
	"errors"
	"strings"
)

// Define a struct to hold the token sanctum data (plain text and hashed text)
type NewToken struct {
	PlainText  string `json:"plain_text"`
	HashedText string `json:"hashed_text"`
}

// GetPlainText returns the plain text
// ModelID.PlainText (e.g. 1|plain_text)
func (nt *NewToken) GetPlainText(ID string) string {
	return ID + "|" + nt.PlainText
}

type TokenSanctum struct {
	Crypto ICryptoSanctum
}

type ITokenSanctum interface {
	CreateToken() (*NewToken, error)
	SplitToken(token string) (string, string, error)
}

// NewTokenSanctum is a function to create a new instance of TokenSanctum
func NewTokenSanctum(
	crypto ICryptoSanctum,
) ITokenSanctum {
	return &TokenSanctum{
		Crypto: crypto,
	}
}

// CreateToken is a function to create a new token sanctum
func (t *TokenSanctum) CreateToken() (*NewToken, error) {
	plainText, err := GenerateRandomString(40) // 40 characters
	if err != nil {
		return nil, err
	}
	hashedText := t.Crypto.SHA256(plainText)
	return &NewToken{
		PlainText:  plainText,
		HashedText: hashedText,
	}, nil
}

// SplitToken is a function to split a token into plain text and hashed text
func (t *TokenSanctum) SplitToken(token string) (string, string, error) {
	parts := strings.Split(token, "|")
	if len(parts) != 2 {
		return "", "", errors.New("invalid token")
	}
	return parts[0], t.Crypto.SHA256(parts[1]), nil
}

var _ ITokenSanctum = (*TokenSanctum)(nil)
