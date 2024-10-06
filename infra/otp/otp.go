package otp

import (
	"shareway/util"

	"github.com/twilio/twilio-go"
)

func NewOTPClient(cfg util.Config) *twilio.RestClient {
	client := twilio.NewRestClientWithParams(
		twilio.ClientParams{
			Username: cfg.TwilioAccountSID,
			Password: cfg.TwilioAuthToken,
		},
	)
	return client
}
