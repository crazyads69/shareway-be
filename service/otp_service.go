package service

import (
	"context"
	"errors"
	"fmt"
	"shareway/infra/otp"
	"shareway/util"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/twilio/twilio-go"
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
	redisClient  *redis.Client
}

// NewOTPService creates a new OTPService instance
func NewOTPService(cfg util.Config, redisClient *redis.Client) IOTPService {
	client := otp.NewOTPClient(cfg)
	return &OTPService{
		twilioClient: client,
		cfg:          cfg,
		redisClient:  redisClient,
	}
}

// SendOTP sends an OTP to the specified phone number
func (s *OTPService) SendOTP(ctx context.Context, phoneNumber string) (string, error) {

	// Check the number of OTPs sent to the phone number
	sendCountKey := fmt.Sprintf("otp:send_count:%s", phoneNumber)
	sendCount, err := s.redisClient.Get(ctx, sendCountKey).Int()
	if err != nil && err != redis.Nil {
		return "", err
	}

	if sendCount >= s.cfg.MaxOTPSendCount {
		return "", fmt.Errorf("You have reached the maximum number of OTP requests. Please try again later")
	}

	// Check cooldown time between OTP sends
	cooldownKey := fmt.Sprintf("otp:cooldown:%s", phoneNumber)
	cooldown, err := s.redisClient.TTL(ctx, cooldownKey).Result()
	if err != nil && err != redis.Nil {
		return "", err
	}

	if cooldown > 0 {
		return "", fmt.Errorf("Please wait %d seconds before requesting a new OTP", int(cooldown.Seconds()))
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
	pipe := s.redisClient.Pipeline()
	pipe.Incr(ctx, sendCountKey)
	pipe.Expire(ctx, sendCountKey, time.Second*time.Duration(s.cfg.OTPSendCountDuration))
	pipe.Set(ctx, cooldownKey, "cooldown", time.Second*time.Duration(s.cfg.OtpCooldownDuration))
	_, err = pipe.Exec(ctx)
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

	// Delete all related keys when OTP is verified successfully
	sendCountKey := fmt.Sprintf("otp:send_count:%s", phoneNumber)
	cooldownKey := fmt.Sprintf("otp:cooldown:%s", phoneNumber)

	// Delete keys in a transaction to ensure atomicity
	pipe := s.redisClient.Pipeline()
	pipe.Del(ctx, sendCountKey)
	pipe.Del(ctx, cooldownKey)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

// Ensure OTPService implements IOTPService
var _ IOTPService = (*OTPService)(nil)
