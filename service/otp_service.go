package service

import (
	"errors"
	"shareway/infra/otp"
	"shareway/util"

	"github.com/twilio/twilio-go"
	twilioApi "github.com/twilio/twilio-go/rest/verify/v2"
)

// formatPhoneNumber converts the phone number to E.164 format
// func formatPhoneNumber(phoneNumber string) string {
// 	return "+84" + phoneNumber
// }

// IOTPService defines the interface for OTP operations
type IOTPService interface {
	SendOTP(phoneNumber string) (string, error)
	VerifyOTP(phoneNumber, code string) error
}

// OTPService implements IOTPService and handles OTP-related operations
type OTPService struct {
	twilioClient *twilio.RestClient
	cfg          util.Config
}

// NewOTPService creates a new OTPService instance
func NewOTPService(cfg util.Config) IOTPService {
	client := otp.NewOTPClient(cfg)
	return &OTPService{
		twilioClient: client,
		cfg:          cfg,
	}
}

// SendOTP sends an OTP to the specified phone number
func (s *OTPService) SendOTP(phoneNumber string) (string, error) {
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

	return *resp.Sid, nil
}

// VerifyOTP verifies the OTP for the given phone number
func (s *OTPService) VerifyOTP(phoneNumber, code string) error {
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

	return nil
}

// Ensure OTPService implements IOTPService
var _ IOTPService = (*OTPService)(nil)
