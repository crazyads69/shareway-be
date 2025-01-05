package repository

import (
	"errors"
	"fmt"
	"shareway/infra/db/migration"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// IAuthRepository defines the interface for authentication-related database operations
type IAuthRepository interface {
	UserExistsByPhone(phoneNumber string) (bool, error)
	GetUserIDByPhone(phoneNumber string) (uuid.UUID, error)
	ActivateUser(phoneNumber string) error
	GetUserByPhone(phoneNumber string) (migration.User, error)
	SaveCCCDInfo(cccdEncrypted string, userID uuid.UUID) error
	VerifyUser(phoneNumber string) error
	SaveSession(phoneNumber string, accessToken string, refreshToken string, userID uuid.UUID) error
	UserExistsByEmail(email string) (bool, error)
	CreateUser(phoneNumber string, fullName string, email string) (uuid.UUID, error)
	GetUserByEmail(email string) (migration.User, error)
	UpdateSession(accessToken string, userID uuid.UUID, refreshToken string) error
	RevokeToken(userID uuid.UUID, refreshToken string) error
	GetUserByID(userID uuid.UUID) (migration.User, error)
	RegisterDeviceToken(userID uuid.UUID, deviceToken string) error
	DeleteUser(phoneNumber string) error
	UpdateUserProfile(userID uuid.UUID, fullName string, email string, gender string) error
	UpdateAvatar(userID uuid.UUID, avatarURL string) error
}

// AuthRepository implements IAuthRepository
type AuthRepository struct {
	db *gorm.DB
}

// NewAuthRepository creates a new instance of AuthRepository
func NewAuthRepository(db *gorm.DB) IAuthRepository {
	return &AuthRepository{db: db}
}

// UserExistsByPhone checks if a user exists with the given phone number
func (r *AuthRepository) UserExistsByPhone(phoneNumber string) (bool, error) {
	var exists bool
	err := r.db.Model(&migration.User{}).
		Select("1").
		Where("phone_number = ?", phoneNumber).
		Limit(1).
		Find(&exists).
		Error
	return exists, err
}

// GetUserIDByPhone retrieves the user ID associated with the given phone number
func (r *AuthRepository) GetUserIDByPhone(phoneNumber string) (uuid.UUID, error) {
	var user migration.User
	err := r.db.Model(&migration.User{}).
		Select("id").
		Where("phone_number = ?", phoneNumber).
		First(&user).
		Error
	return user.ID, err
}

// ActivateUser updates the user status to activated
func (r *AuthRepository) ActivateUser(phoneNumber string) error {
	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Model(&migration.User{}).
		Where("phone_number = ?", phoneNumber).
		Update("is_activated", true).
		Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// GetUserByPhone retrieves the user associated with the given phone number
func (r *AuthRepository) GetUserByPhone(phoneNumber string) (migration.User, error) {
	var user migration.User
	err := r.db.Where("phone_number = ?", phoneNumber).First(&user).Error
	return user, err
}

// SaveCCCDInfo saves the encrypted CCCD information to the database
func (r *AuthRepository) SaveCCCDInfo(cccdEncrypted string, userID uuid.UUID) error {
	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Model(&migration.User{}).Where("id = ?", userID).Update("cccd_number", cccdEncrypted).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// VerifyUser updates the user status to verified
func (r *AuthRepository) VerifyUser(phoneNumber string) error {
	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Model(&migration.User{}).
		Where("phone_number = ?", phoneNumber).
		Update("is_verified", true).
		Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// SaveSession saves the access token and refresh token to the database
func (r *AuthRepository) SaveSession(phoneNumber string, accessToken string, refreshToken string, userID uuid.UUID) error {
	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	token := migration.PasetoToken{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		UserID:       userID,
	}
	if err := tx.Create(&token).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// UserExistsByEmail checks if a user exists with the given email
func (r *AuthRepository) UserExistsByEmail(email string) (bool, error) {
	var exists bool
	err := r.db.Model(&migration.User{}).
		Select("1").
		Where("email = ?", email).
		Limit(1).
		Find(&exists).
		Error
	return exists, err

}

// CreateUser creates a new user with the given phone number, full name, and email
func (r *AuthRepository) CreateUser(phoneNumber, fullName, email string) (uuid.UUID, error) {
	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	user := migration.User{
		PhoneNumber: phoneNumber,
		FullName:    fullName,
	}
	if email != "" {
		user.Email = email
	}

	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		return uuid.Nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Update user avatar
	avatarURL := fmt.Sprintf("https://api.multiavatar.com/%s.png", user.ID)
	if err := tx.Model(&migration.User{}).Where("id = ?", user.ID).Update("avatar_url", avatarURL).Error; err != nil {
		tx.Rollback()
		return uuid.Nil, fmt.Errorf("failed to update user avatar: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return uuid.Nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return user.ID, nil
}

// GetUserByEmail retrieves a user by their email address
func (r *AuthRepository) GetUserByEmail(email string) (migration.User, error) {
	var user migration.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return migration.User{}, fmt.Errorf("user with email %s not found", email)
		}
		return migration.User{}, fmt.Errorf("failed to fetch user: %w", err)
	}
	return user, nil
}

// UpdateSession updates the access token for the given user ID
func (r *AuthRepository) UpdateSession(accessToken string, userID uuid.UUID, refreshToken string) error {
	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// First, get the current token record with the given user ID and refresh token
	var token migration.PasetoToken
	if err := tx.Where("user_id = ? AND refresh_token = ?", userID, refreshToken).First(&token).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("no token found for user")
		}
		return fmt.Errorf("failed to fetch token: %w", err)
	}
	// If the token is already revoked, return an error message meaning the user has to log in again
	if token.Revoke {
		tx.Rollback()
		return fmt.Errorf("token has been revoked: please log in again")
	}
	// Check if refresh_turns is 3 or more
	if token.RefreshTurns >= 3 {
		// If so, revoke the token
		if err := tx.Model(&token).Updates(map[string]interface{}{
			"revoke": true,
		}).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to revoke token: %w", err)
		}
		if err := tx.Commit().Error; err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}
		return fmt.Errorf("token has been revoked due to maximum refresh limit: please log in again")
	} else {
		// Otherwise, update the access token and increment refresh_turns
		if err := tx.Model(&token).Updates(map[string]interface{}{
			"access_token":  accessToken,
			"refresh_turns": token.RefreshTurns + 1,
		}).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update token: %w", err)
		}
	}

	return tx.Commit().Error
}

// RevokeToken revokes the given refresh token
func (r *AuthRepository) RevokeToken(userID uuid.UUID, refreshToken string) error {
	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// First, get the current token record
	var token migration.PasetoToken
	if err := tx.Where("user_id = ? AND refresh_token = ?", userID, refreshToken).First(&token).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("no token found for user")
		}
		return fmt.Errorf("failed to fetch token: %w", err)
	}

	// Check if the token is already revoked
	if token.Revoke {
		tx.Rollback()
		return nil // Token already revoked, consider this a success
	}

	// Revoke the token
	if err := tx.Model(&token).Updates(map[string]interface{}{
		"revoke":        true,
		"refresh_turns": 3, // Set refresh_turns to 3 to prevent further refreshes
	}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to revoke token: %w", err)
	}

	return tx.Commit().Error
}

// GetUserByID retrieves the user associated with the given user ID
func (r *AuthRepository) GetUserByID(userID uuid.UUID) (migration.User, error) {
	// Get the user record with the given user ID
	var user migration.User
	err := r.db.First(&user, userID).Error
	return user, err
}

// RegisterDeviceToken registers the device token for the given user ID in the database
func (r *AuthRepository) RegisterDeviceToken(userID uuid.UUID, deviceToken string) error {
	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Update the device token for the user with the given user ID
	result := tx.Model(&migration.User{}).
		Where("id = ?", userID).
		Update("device_token", deviceToken)

	if result.Error != nil {
		tx.Rollback()
		return result.Error
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		return errors.New("user not found")
	}

	return tx.Commit().Error
}

// DeleteUser delete user from given phone number in db
func (r *AuthRepository) DeleteUser(phoneNumber string) error {
	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Check if the user exists
	var user migration.User
	if err := tx.Where("phone_number = ?", phoneNumber).First(&user).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return err
	}

	// Delete related records in paseto_tokens table
	if err := tx.Where("user_id = ?", user.ID).Delete(&migration.PasetoToken{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Delete related records in other tables
	// OTP
	if err := tx.Where("user_id = ?", user.ID).Delete(&migration.OTP{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Delete Rides associated with user's RideOffers and RideRequests
	var rideOffers []migration.RideOffer
	var rideRequests []migration.RideRequest
	if err := tx.Where("user_id = ?", user.ID).Find(&rideOffers).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Where("user_id = ?", user.ID).Find(&rideRequests).Error; err != nil {
		tx.Rollback()
		return err
	}

	var rideOfferIDs []uuid.UUID
	var rideRequestIDs []uuid.UUID
	for _, offer := range rideOffers {
		rideOfferIDs = append(rideOfferIDs, offer.ID)
	}
	for _, request := range rideRequests {
		rideRequestIDs = append(rideRequestIDs, request.ID)
	}

	if err := tx.Where("ride_offer_id IN ? OR ride_request_id IN ?", rideOfferIDs, rideRequestIDs).Delete(&migration.Ride{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// RideRequests
	if err := tx.Where("user_id = ?", user.ID).Delete(&migration.RideRequest{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// RideOffers
	if err := tx.Where("user_id = ?", user.ID).Delete(&migration.RideOffer{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Vehicles (now safe to delete after RideOffers are deleted)
	if err := tx.Where("user_id = ?", user.ID).Delete(&migration.Vehicle{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Notifications
	if err := tx.Where("user_id = ?", user.ID).Delete(&migration.Notification{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// FavoriteLocations
	if err := tx.Where("user_id = ?", user.ID).Delete(&migration.FavoriteLocation{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Chats (both sent and received)
	if err := tx.Where("sender_id = ? OR receiver_id = ?", user.ID, user.ID).Delete(&migration.Chat{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Ratings (both given and received)
	if err := tx.Where("rater_id = ? OR ratee_id = ?", user.ID, user.ID).Delete(&migration.Rating{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Finally, delete the user
	if err := tx.Delete(&user).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// UpdateUser updates the user with the given user ID
func (r *AuthRepository) UpdateUserProfile(userID uuid.UUID, fullName, email, gender string) error {
	tx := r.db.Begin()
	// Defer a function to handle rollback or commit
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create a map to hold the fields to update
	updates := map[string]interface{}{
		"gender":    gender,
		"full_name": fullName,
	}

	// Only include email in updates if it's not empty
	if email != "" {
		updates["email"] = email
	}

	// Update the user record with the given user ID within the transaction
	result := tx.Model(&migration.User{}).Where("id = ?", userID).Updates(updates)

	if result.Error != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update user profile: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		return errors.New("user not found")
	}

	return tx.Commit().Error
}

// UpdateAvatar updates the user avatar with the given user ID
func (r *AuthRepository) UpdateAvatar(userID uuid.UUID, avatarURL string) error {
	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Update the user avatar URL with the given user ID
	result := tx.Model(&migration.User{}).
		Where("id = ?", userID).
		Update("avatar_url", avatarURL)

	if result.Error != nil {
		tx.Rollback()
		return result.Error
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		return errors.New("user not found")
	}

	return tx.Commit().Error
}

// Ensure AuthRepository implements IAuthRepository
var _ IAuthRepository = (*AuthRepository)(nil)
