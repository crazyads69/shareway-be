package service

import (
	"context"
	"errors"
	"fmt"

	"shareway/infra/otp"
	"shareway/repository"
	"shareway/util"

	twilioApi "github.com/twilio/twilio-go/rest/verify/v2"
)

// formatPhoneNumber converts the phone number to E.164 format
// func formatPhoneNumber(phoneNumber string) string {
// 	return "+84" + phoneNumber
// }

// IOTPService defines the interface for OTP operations
type IOTPService interface {
	SendOTP(ctx context.Context, phoneNumber string) (string, error)
	VerifyOTP(ctx context.Context, phoneNumber, code string) error
}

// OTPService implements IOTPService and handles OTP-related operations
type OTPService struct {
	twilioClient *twilio.RestClient
	cfg          util.Config
	repo         repository.IOTPRepository
}

// NewOTPService creates a new OTPService instance
func NewOTPService(cfg util.Config, repo repository.IOTPRepository) IOTPService {
	client := otp.NewOTPClient(cfg)
	return &OTPService{
		twilioClient: client,
		cfg:          cfg,
		repo:         repo,
	}
}

// SendOTP sends an OTP to the specified phone number
func (s *OTPService) SendOTP(ctx context.Context, phoneNumber string) (string, error) {

	sendCount, err := s.repo.CheckSendCount(ctx, phoneNumber)
	if err != nil {
		return "", err
	}

	if sendCount >= s.cfg.MaxOTPSendCount {
		return "", fmt.Errorf("You have reached the maximum number of OTP requests. Please try again later")
	}

	// Check cooldown time between OTP sends
	cooldown, err := s.repo.CheckCooldown(ctx, phoneNumber)
	if err != nil {
		return "", err
	}

	if cooldown > 0 {
		return "", fmt.Errorf("please wait %d seconds before requesting a new OTP", int(cooldown.Seconds()))
	}

	// Generate and send OTP
	params := &twilioApi.CreateVerificationParams{}
	params.SetTo(phoneNumber) // Phone number in E.164 format already
	params.SetChannel("sms")

	resp, err := s.twilioClient.VerifyV2.CreateVerification(s.cfg.TwilioServiceSID, params)
	if err != nil {
		return "", err
	}

	if resp.Sid == nil {
		return "", errors.New("failed to get verification SID")
	}

	// Increment the send count and set cooldown
	err = s.repo.IncrementSendCount(ctx, phoneNumber, s.cfg.OTPSendCountDuration)
	if err != nil {
		return "", err
	}

	err = s.repo.SetCooldown(ctx, phoneNumber, s.cfg.OtpCooldownDuration)
	if err != nil {
		return "", err
	}

	return *resp.Sid, nil
}

// VerifyOTP verifies the OTP for the given phone number
func (s *OTPService) VerifyOTP(ctx context.Context, phoneNumber, code string) error {
	params := &twilioApi.CreateVerificationCheckParams{}
	params.SetTo(phoneNumber) // Phone number in E.164 format already
	params.SetCode(code)

	resp, err := s.twilioClient.VerifyV2.CreateVerificationCheck(s.cfg.TwilioServiceSID, params)
	if err != nil {
		return err
	}

	if resp.Status == nil {
		return errors.New("failed to get verification status")
	}

	if *resp.Status != "approved" {
		return errors.New("OTP verification failed")
	}

	// Delete keys in a transaction to ensure atomicity
	err = s.repo.ClearOTPData(ctx, phoneNumber)
	if err != nil {
		return err
	}

	return nil
}

// Ensure OTPService implements IOTPService
var _ IOTPService = (*OTPService)(nil)
