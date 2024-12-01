package repository

import (
	"shareway/infra/db/migration"
	"shareway/schemas"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type IPNRepository struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewIPNRepository(db *gorm.DB, redis *redis.Client) IIPNRepository {
	return &IPNRepository{
		db:    db,
		redis: redis,
	}
}

type IIPNRepository interface {
	GetUserByPartnerClientID(partnerClientID string) (migration.User, error)
	UpdateUserMoMoToken(userID uuid.UUID, token schemas.DecodedToken) error
	StoreCallbackToken(token string, userID uuid.UUID) error
	StoreTransID(transID int64, rideRequestID uuid.UUID) error
}

func (p *IPNRepository) GetUserByPartnerClientID(partnerClientID string) (migration.User, error) {
	var user migration.User
	newPartnerClientID, err := uuid.Parse(partnerClientID)
	if err != nil {
		return migration.User{}, err
	}
	if err := p.db.Where("id = ?", newPartnerClientID).First(&user).Error; err != nil {
		return user, err
	}

	return user, nil
}

func (p *IPNRepository) UpdateUserMoMoToken(userID uuid.UUID, token schemas.DecodedToken) error {
	var user migration.User
	if err := p.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return err
	}

	user.MoMoRecurringToken = token.Value
	if err := p.db.Save(&user).Error; err != nil {
		return err
	}
	return nil
}

func (p *IPNRepository) StoreCallbackToken(token string, userID uuid.UUID) error {
	var user migration.User
	if err := p.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return err
	}

	user.MoMoCallbackToken = token
	if err := p.db.Save(&user).Error; err != nil {
		return err
	}
	return nil
}

func (p *IPNRepository) StoreTransID(transID int64, rideRequestID uuid.UUID) error {
	// Store IPN transid to db with ride request ID from extra data
	var rideRequest migration.RideRequest
	if err := p.db.Where("id = ?", rideRequestID).First(&rideRequest).Error; err != nil {
		return err
	}

	rideRequest.MomoTransID = transID
	if err := p.db.Save(&rideRequest).Error; err != nil {
		return err
	}
	return nil
}
