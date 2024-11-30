package service

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"shareway/infra/ws"
	"shareway/repository"
	"shareway/schemas"
	"shareway/util"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type PaymentService struct {
	repo repository.IPaymentRepository
	hub  *ws.Hub
	cfg  util.Config
}

type IPaymentService interface {
	LinkMomoWallet(userID uuid.UUID) (schemas.LinkWalletResponse, error)
}

func NewPaymentService(repo repository.IPaymentRepository, hub *ws.Hub, cfg util.Config) IPaymentService {
	return &PaymentService{
		repo: repo,
		hub:  hub,
		cfg:  cfg,
	}
}
func (p *PaymentService) LinkMomoWallet(userID uuid.UUID) (schemas.LinkWalletResponse, error) {
	log.Info().Msg("Starting LinkMomoWallet process")

	// Generate request ID for linking wallet
	requestID := uuid.New().String()
	log.Info().Str("requestID", requestID).Msg("Generated request ID")

	// Store request ID to database
	err := p.repo.StoreRequestID(requestID, userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to store request ID in database")
		return schemas.LinkWalletResponse{}, err
	}
	log.Info().Msg("Stored request ID in database")

	// Build request signature
	var rawSignature bytes.Buffer
	rawSignature.WriteString("accessKey=")
	rawSignature.WriteString(p.cfg.MomoAccessKey)
	rawSignature.WriteString("&amount=0")
	rawSignature.WriteString("&extraData=")
	rawSignature.WriteString("")
	rawSignature.WriteString("&ipnUrl=")
	rawSignature.WriteString(p.cfg.MomoPaymentNotifyURL)
	rawSignature.WriteString("&orderId=")
	rawSignature.WriteString(requestID)
	rawSignature.WriteString("&orderInfo=")
	rawSignature.WriteString("Link wallet to user account")
	rawSignature.WriteString("&partnerClientId=")
	rawSignature.WriteString(userID.String())
	rawSignature.WriteString("&partnerCode=")
	rawSignature.WriteString(p.cfg.MomoPartnerCode)
	rawSignature.WriteString("&redirectUrl=")
	rawSignature.WriteString("")
	rawSignature.WriteString("&requestId=")
	rawSignature.WriteString(requestID)
	rawSignature.WriteString("&requestType=")
	rawSignature.WriteString("linkWallet")

	log.Debug().Str("rawSignature", rawSignature.String()).Msg("Built raw signature")

	// Sign request
	hmac := hmac.New(sha256.New, []byte(p.cfg.MomoSecretKey))
	hmac.Write(rawSignature.Bytes())
	signature := hex.EncodeToString(hmac.Sum(nil))
	log.Debug().Str("signature", signature).Msg("Generated signature")

	// Build request payload
	payload := schemas.LinkWalletRequest{
		PartnerCode:     p.cfg.MomoPartnerCode,
		AccessKey:       p.cfg.MomoAccessKey,
		RequestID:       requestID,
		Amount:          0,
		OrderID:         requestID,
		OrderInfo:       "Link wallet to user account",
		RedirectURL:     "",
		IpnURL:          p.cfg.MomoPaymentNotifyURL,
		PartnerClientID: userID.String(),
		ExtraData:       "",
		RequestType:     "linkWallet",
		Lang:            "vi",
		Signature:       signature,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal payload")
		return schemas.LinkWalletResponse{}, err
	}
	log.Debug().RawJSON("payload", jsonPayload).Msg("Prepared request payload")

	// Send request to MoMo API
	url := fmt.Sprintf("%s/%s", p.cfg.MomoPaymentURL, "create")
	log.Info().Str("url", url).Msg("Sending request to MoMo API")

	start := time.Now()
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Error().Err(err).Msg("Failed to send request to MoMo API")
		return schemas.LinkWalletResponse{}, err
	}
	defer resp.Body.Close()

	duration := time.Since(start)
	log.Info().Dur("duration", duration).Int("statusCode", resp.StatusCode).Msg("Received response from MoMo API")
	// Read response from MoMo API
	var response schemas.LinkWalletResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		log.Error().Err(err).Msg("Failed to decode response from MoMo API")
		return schemas.LinkWalletResponse{}, err
	}

	log.Debug().Interface("response", response).Msg("Decoded response from MoMo API")

	// Check if response is successful by checking result code
	if response.ResultCode != 0 {
		log.Error().Int("resultCode", response.ResultCode).Str("message", response.Message).Msg("Failed to link wallet")
		return schemas.LinkWalletResponse{}, fmt.Errorf("failed to link wallet: %s", response.Message)
	}

	log.Info().Msg("Successfully linked wallet")
	return response, nil
}
