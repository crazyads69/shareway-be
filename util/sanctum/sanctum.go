package sanctum

import (
	"fmt"
	"shareway/infra/db/migration"
	"strconv"
	"time"

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
	Token ITokenSanctum
	db    *gorm.DB
}

func NewSanctumToken(token ITokenSanctum, db *gorm.DB) *SanctumToken {
	return &SanctumToken{
		Token: token,
		db:    db,
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
	// Split the token
	tokenID, plainText, err := st.Token.SplitToken(token)
	if err != nil {
		return nil, err
	}

	// Find the token from the database
	var dbToken migration.SanctumToken
	if err := st.db.Where("token = ? AND is_revoked = ?", plainText, false).
		First(&dbToken, tokenID).Error; err != nil {
		return nil, err
	}
	// Verify the token expiration
	if dbToken.ExpiredAt.Before(time.Now().UTC()) {
		return nil, fmt.Errorf("token has expired")
	}

	// Return the token payload
	return &SanctumTokenPayload{
		AdminID:   dbToken.AdminID,
		TokenID:   dbToken.ID,
		ExpiredAt: dbToken.ExpiredAt,
		CreatedAt: dbToken.CreatedAt,
		Ability:   dbToken.Ability,
	}, nil
}
