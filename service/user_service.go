package service

import (
	"mime/multipart"
	"shareway/infra/db/migration"
	"shareway/infra/fpt"
	"shareway/repository"
	"shareway/schemas"
	"shareway/util"
	"shareway/util/token"
	"time"

	"github.com/google/uuid"
)

// IUsersService defines the interface for user-related business logic operations
type IUsersService interface {
	UserExistsByPhone(phoneNumber string) (bool, error)
	CreateUserByPhone(phoneNumber, fullName string) (uuid.UUID, string, error)
	GetUserIDByPhone(phoneNumber string) (uuid.UUID, error)
	ActivateUser(phoneNumber string) error
	GetUserByPhone(phoneNumber string) (migration.User, error)
	VerifyCCCD(image *multipart.FileHeader) (*fpt.CCCDInfo, error)
	EncryptAndSaveCCCDInfo(cccdInfo *fpt.CCCDInfo, userID uuid.UUID) error
	VerifyUser(phoneNumber string) error
	CreateSession(phoneNumber string, userID uuid.UUID) (migration.User, string, string, error)
	UserExistsByEmail(email string) (bool, error)
	CreateUser(phoneNumber string, fullName string, email string) (uuid.UUID, error)
	GetUserByEmail(email string) (migration.User, error)
	ValidateRefreshToken(refreshToken string) (schemas.Payload, error)
	RefreshNewToken(phoneNumber string, userID uuid.UUID) (string, error)
	UpdateSession(accessToken string, userID uuid.UUID) error
	RevokeToken(userID uuid.UUID, refreshToken string) error
}

// UsersService implements IUsersService and handles user-related business logic
type UsersService struct {
	repo      repository.IAuthRepository
	encryptor util.IEncryptor
	fptReader *fpt.FPTReader
	maker     *token.PasetoMaker
	cfg       util.Config
}

// NewUsersService creates a new instance of UsersService
func NewUsersService(repo repository.IAuthRepository, encryptor util.IEncryptor, fptReader *fpt.FPTReader, maker *token.PasetoMaker, cfg util.Config) IUsersService {
	return &UsersService{
		repo:      repo,
		encryptor: encryptor,
		fptReader: fptReader,
		maker:     maker,
		cfg:       cfg,
	}
}

// UserExistsByPhone checks if a user exists with the given phone number
func (s *UsersService) UserExistsByPhone(phoneNumber string) (bool, error) {
	return s.repo.UserExistsByPhone(phoneNumber)
}

// CreateUserByPhone creates a new user with the given phone number and full name
func (s *UsersService) CreateUserByPhone(phoneNumber, fullName string) (uuid.UUID, string, error) {
	return s.repo.CreateUserByPhone(phoneNumber, fullName)
}

// GetUserIDByPhone retrieves the user ID associated with the given phone number
func (s *UsersService) GetUserIDByPhone(phoneNumber string) (uuid.UUID, error) {
	return s.repo.GetUserIDByPhone(phoneNumber)
}

// ActivateUser activates the user account associated with the given phone number
func (s *UsersService) ActivateUser(phoneNumber string) error {
	return s.repo.ActivateUser(phoneNumber)
}

// GetUserByPhone retrieves the user associated with the given phone number
func (s *UsersService) GetUserByPhone(phoneNumber string) (migration.User, error) {
	return s.repo.GetUserByPhone(phoneNumber)
}

// VerifyCCCD sends an image to FPT AI for verification and returns the extracted information
func (s *UsersService) VerifyCCCD(image *multipart.FileHeader) (*fpt.CCCDInfo, error) {
	return s.fptReader.VerifyImageWithFPTAI(image)
}

// EncryptAndSaveCCCDInfo encrypts and saves the CCCD ID information for the given user ID
func (s *UsersService) EncryptAndSaveCCCDInfo(cccdInfo *fpt.CCCDInfo, userID uuid.UUID) error {
	encryptedInfo, err := s.encryptor.Encrypt(cccdInfo.ID)
	if err != nil {
		return err
	}

	return s.repo.SaveCCCDInfo(encryptedInfo, userID)
}

// VerifyUser verifies the user account associated with the given phone number
func (s *UsersService) VerifyUser(phoneNumber string) error {
	return s.repo.VerifyUser(phoneNumber)
}

// CreateSession creates a new session for the user with the given user ID
// Access token and refresh token are returned
func (s *UsersService) CreateSession(phoneNumber string, userID uuid.UUID) (migration.User, string, string, error) {
	// First create access token and refresh token
	accessToken, err := s.maker.CreateToken(phoneNumber, userID, time.Duration(s.cfg.AccessTokenExpiredDuration))
	if err != nil {
		return migration.User{}, "", "", err
	}

	refreshToken, err := s.maker.CreateToken(phoneNumber, userID, time.Duration(s.cfg.RefreshTokenExpiredDuration))
	if err != nil {
		return migration.User{}, "", "", err
	}

	// Save the tokens to the database
	err = s.repo.SaveSession(phoneNumber, accessToken, refreshToken, userID)
	if err != nil {
		return migration.User{}, "", "", err
	}

	// Return the user
	user, err := s.GetUserByPhone(phoneNumber)
	if err != nil {
		return migration.User{}, "", "", err
	}

	return user, accessToken, refreshToken, nil
}

// UserExistsByEmail checks if a user exists with the given email
func (s *UsersService) UserExistsByEmail(email string) (bool, error) {
	return s.repo.UserExistsByEmail(email)
}

// CreateUser creates a new user with the given phone number, full name, and email
func (s *UsersService) CreateUser(phoneNumber, fullName, email string) (uuid.UUID, error) {
	return s.repo.CreateUser(phoneNumber, fullName, email)
}

// GetUserByEmail retrieves the user associated with the given email
func (s *UsersService) GetUserByEmail(email string) (migration.User, error) {
	return s.repo.GetUserByEmail(email)
}

// ValidateRefreshToken validates the given refresh token and returns the phone number
func (s *UsersService) ValidateRefreshToken(refreshToken string) (schemas.Payload, error) {
	payload, err := s.maker.VerifyToken(refreshToken)
	if err != nil {
		return schemas.Payload{}, err
	}
	return *payload, nil
}

// RefreshNewToken refreshes the access token for the given phone number and user ID
func (s *UsersService) RefreshNewToken(phoneNumber string, userID uuid.UUID) (string, error) {
	return s.maker.CreateToken(phoneNumber, userID, time.Duration(s.cfg.AccessTokenExpiredDuration))
}

// UpdateSession updates the access token for the given user ID
func (s *UsersService) UpdateSession(accessToken string, userID uuid.UUID) error {
	return s.repo.UpdateSession(accessToken, userID)
}

// RevokeToken revokes the refresh token for the given user ID
func (s *UsersService) RevokeToken(userID uuid.UUID, refreshToken string) error {
	return s.repo.RevokeToken(userID, refreshToken)
}

// Ensure UsersService implements IUsersService
var _ IUsersService = (*UsersService)(nil)
