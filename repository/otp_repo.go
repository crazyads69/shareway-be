package repository

import (
	"context"
	"fmt"
	"time"

	"shareway/util"

	"github.com/redis/go-redis/v9"
)

type IOTPRepository interface {
	CheckSendCount(ctx context.Context, phoneNumber string) (int, error)
	CheckCooldown(ctx context.Context, phoneNumber string) (time.Duration, error)
	IncrementSendCount(ctx context.Context, phoneNumber string, duration int) error
	SetCooldown(ctx context.Context, phoneNumber string, duration int) error
	ClearOTPData(ctx context.Context, phoneNumber string) error
}

type OTPRepository struct {
	redisClient *redis.Client
	cfg         util.Config
}

func NewOTPRepository(redisClient *redis.Client, cfg util.Config) IOTPRepository {
	return &OTPRepository{
		redisClient: redisClient,
		cfg:         cfg,
	}
}

// CheckSendCount checks the number of OTPs sent to the phone number
func (r *OTPRepository) CheckSendCount(ctx context.Context, phoneNumber string) (int, error) {
	// Create a key to store the number of OTPs sent to the phone number
	sendCountKey := fmt.Sprintf("otp:send_count:%s", phoneNumber)
	// Get the number of OTPs sent to the phone number
	sendCount, err := r.redisClient.Get(ctx, sendCountKey).Int()
	if err == redis.Nil {
		return 0, nil
	}
	return sendCount, err
}

// CheckCooldown checks the cooldown time between OTP sends
func (r *OTPRepository) CheckCooldown(ctx context.Context, phoneNumber string) (time.Duration, error) {
	cooldownKey := fmt.Sprintf("otp:cooldown:%s", phoneNumber)
	// Get the remaining time until the cooldown expires
	return r.redisClient.TTL(ctx, cooldownKey).Result()
}

// IncrementSendCount increments the number of OTPs sent to the phone number
func (r *OTPRepository) IncrementSendCount(ctx context.Context, phoneNumber string, duration int) error {
	sendCountKey := fmt.Sprintf("otp:send_count:%s", phoneNumber)
	pipe := r.redisClient.Pipeline()
	pipe.Incr(ctx, sendCountKey)
	pipe.Expire(ctx, sendCountKey, time.Second*time.Duration(duration))
	_, err := pipe.Exec(ctx)
	return err
}

// SetCooldown sets the cooldown time between OTP sends
func (r *OTPRepository) SetCooldown(ctx context.Context, phoneNumber string, duration int) error {
	cooldownKey := fmt.Sprintf("otp:cooldown:%s", phoneNumber)
	return r.redisClient.Set(ctx, cooldownKey, "cooldown", time.Second*time.Duration(duration)).Err()
}

// ClearOTPData clears the OTP data for the phone number
func (r *OTPRepository) ClearOTPData(ctx context.Context, phoneNumber string) error {
	sendCountKey := fmt.Sprintf("otp:send_count:%s", phoneNumber)
	cooldownKey := fmt.Sprintf("otp:cooldown:%s", phoneNumber)

	pipe := r.redisClient.Pipeline()
	pipe.Del(ctx, sendCountKey)
	pipe.Del(ctx, cooldownKey)
	_, err := pipe.Exec(ctx)
	return err
}

// Make sure OTPRepository implements IOTPRepository
var _ IOTPRepository = (*OTPRepository)(nil)
