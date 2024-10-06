package service

import (
	"errors"
	"shareway/infra/otp"
	"shareway/util"

	"github.com/twilio/twilio-go"
	twilioApi "github.com/twilio/twilio-go/rest/verify/v2"
)

// OtpService handles OTP-related operations
type OtpService struct {
	twilioClient *twilio.RestClient
	cfg          util.Config
}

// IOTPService defines the interface for OTP operations
type IOTPService interface {
	SendOTP(phoneNumber string) (string, error)
	VerifyOTP(phoneNumber, code string) error
}

// NewOTPService creates a new OTPService instance
func NewOTPService(cfg util.Config) *OtpService {
	client := otp.NewOTPClient(cfg)
	return &OtpService{
		twilioClient: client,
		cfg:          cfg,
	}
}

// SendOTP sends an OTP to the specified phone number
func (s *OtpService) SendOTP(phoneNumber string) (string, error) {
	// Convert the phone number to E.164 format
	newPhoneNumber := "+84" + phoneNumber
	params := &twilioApi.CreateVerificationParams{}
	params.SetTo(newPhoneNumber)
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
func (s *OtpService) VerifyOTP(phoneNumber, code string) error {
	// Convert the phone number to E.164 format
	newPhoneNumber := "+84" + phoneNumber
	params := &twilioApi.CreateVerificationCheckParams{}
	params.SetTo(newPhoneNumber)
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
