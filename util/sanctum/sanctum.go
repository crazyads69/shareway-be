package sanctum

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"shareway/infra/db/migration"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SanctumTokenPayload struct {
	AdminID   uuid.UUID `json:"admin_id"`
	TokenID   int64     `json:"token_id"`
	ExpiredAt time.Time `json:"expired_at"`
	CreatedAt time.Time `json:"created_at"`
	Ability   string    `json:"ability"` // Default is "*" for admin
}

type SanctumToken struct {
	Token  ITokenSanctum
	Crypto ICryptoSanctum
	db     *gorm.DB
}

func NewSanctumToken(token ITokenSanctum, crypto ICryptoSanctum, db *gorm.DB) *SanctumToken {
	return &SanctumToken{
		Token:  token,
		Crypto: crypto,
		db:     db,
	}
}

type ISanctumToken interface {
	CreateSanctumToken(adminID uuid.UUID, duration time.Duration) (string, error)
	VerifySanctumToken(token string) (*SanctumTokenPayload, error)
}

func (st *SanctumToken) CreateSanctumToken(adminID uuid.UUID, duration time.Duration) (string, error) {
	// Create a new token
	newToken, err := st.Token.CreateToken()
	if err != nil {
		return "", err
	}

	// Create the current timestamp
	now := time.Now().UTC()

	// First insert the token to the database to get the token ID
	token := migration.SanctumToken{
		AdminID:   adminID,
		Ability:   "*",
		ExpiredAt: now.Add(duration),
		Token:     newToken.HashedText,
	}
	if err := st.db.Create(&token).Error; err != nil {
		return "", err
	}

	// Return the plain text token
	return newToken.GetPlainText(strconv.FormatInt(token.ID, 10)), nil
}

func (st *SanctumToken) VerifySanctumToken(token string) (*SanctumTokenPayload, error) {
	tokenID, hashedToken, err := st.Token.SplitToken(token)
	if err != nil {
		return nil, err
	}

	var dbToken migration.SanctumToken
	if err := st.db.Where("token = ? AND is_revoked = ?", hashedToken, false).
		First(&dbToken, tokenID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("token not found")
		}
		return nil, err
	}

	// Use constant-time comparison
	if !st.Crypto.SecureCompare(hashedToken, dbToken.Token) {
		return nil, fmt.Errorf("token does not match")
	}

	if dbToken.ExpiredAt.Before(time.Now().UTC()) {
		return nil, fmt.Errorf("token has expired at %s", dbToken.ExpiredAt)
	}

	return &SanctumTokenPayload{
		AdminID:   dbToken.AdminID,
		TokenID:   dbToken.ID,
		ExpiredAt: dbToken.ExpiredAt,
		CreatedAt: dbToken.CreatedAt,
		Ability:   dbToken.Ability,
	}, nil
}

func (st *SanctumToken) InvalidateToken(data *SanctumTokenPayload) error {
	// Set the token to be revoked
	if err := st.db.Model(&migration.SanctumToken{}).
		Where("id = ?", data.TokenID).
		Update("is_revoked", true).Error; err != nil {
		return err
	}

	// Update the expired at to the current time
	if err := st.db.Model(&migration.SanctumToken{}).
		Where("id = ?", data.TokenID).
		Update("expired_at", time.Now().UTC()).Error; err != nil {
		return err
	}

	return nil
}
