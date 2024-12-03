package service

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
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
	LinkMomoWallet(userID uuid.UUID, walletPhoneNumber string) (schemas.LinkWalletResponse, error)
	CheckoutRide(userID uuid.UUID, req schemas.CheckoutRideRequest) error
	encryptRSA(tokenData schemas.TokenData) (string, error)
	RefundRide(userID uuid.UUID, req schemas.RefundMomoRequest) error
}

func NewPaymentService(repo repository.IPaymentRepository, hub *ws.Hub, cfg util.Config) IPaymentService {
	return &PaymentService{
		repo: repo,
		hub:  hub,
		cfg:  cfg,
	}
}
func (p *PaymentService) LinkMomoWallet(userID uuid.UUID, walletPhoneNumber string) (schemas.LinkWalletResponse, error) {
	log.Info().Msg("Starting LinkMomoWallet process")

	// Generate request ID for linking wallet
	requestID := uuid.New().String()
	log.Info().Str("requestID", requestID).Msg("Generated request ID")

	// Store request ID to database
	err := p.repo.StoreRequestID(requestID, userID, walletPhoneNumber)

	// Store the wallet phone number to database
	if err != nil {
		log.Error().Err(err).Msg("Failed to store request ID in database")
		return schemas.LinkWalletResponse{}, err
	}
	log.Info().Msg("Stored request ID in database")

	extraType := schemas.ExtraData{
		Type: "linkWallet",
	}

	// Encode extra data to JSON
	extraData, err := json.Marshal(extraType)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal extra data")
		return schemas.LinkWalletResponse{}, err
	}

	// Encode extra data to base64
	extraDataBase64 := base64.StdEncoding.EncodeToString(extraData)
	// Build request signature
	var rawSignature bytes.Buffer
	rawSignature.WriteString("accessKey=")
	rawSignature.WriteString(p.cfg.MomoAccessKey)
	rawSignature.WriteString("&amount=0")
	rawSignature.WriteString("&extraData=")
	rawSignature.WriteString(extraDataBase64)
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
	rawSignature.WriteString(p.cfg.MomoPaymentRedirectURL)
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
		RedirectURL:     p.cfg.MomoPaymentRedirectURL,
		IpnURL:          p.cfg.MomoPaymentNotifyURL,
		PartnerClientID: userID.String(),
		ExtraData:       extraDataBase64,
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

func (p *PaymentService) CheckoutRide(userID uuid.UUID, req schemas.CheckoutRideRequest) error {
	log.Info().Msg("Starting CheckoutRide process")
	// Get checkout token from user
	user, err := p.repo.GetUserByID(userID)
	if err != nil {
		log.Error().Err(err).Str("userID", userID.String()).Msg("Failed to get user details")
		return fmt.Errorf("failed to get user details: %w", err)
	}

	// Get ride offer details
	rideOffer, err := p.repo.GetRideOfferByID(req.RideOfferID)
	if err != nil {
		log.Error().Err(err).Str("rideOfferID", req.RideOfferID.String()).Msg("Failed to get ride offer details")
		return fmt.Errorf("failed to get ride offer details: %w", err)
	}

	// Generate request ID
	requestID := uuid.New().String()

	// Prepare token data
	tokenData := schemas.TokenData{
		Value:               user.MoMoRecurringToken,
		RequireSecurityCode: false,
	}

	// Encrypt token data RSA
	// Encrypt token data with RSA
	encryptedToken, err := p.encryptRSA(tokenData)
	if err != nil {
		log.Error().Err(err).Msg("Failed to encrypt token data")
		return fmt.Errorf("failed to encrypt token data: %w", err)
	}

	// Prepare extra data
	extraData := schemas.ExtraData{
		Type:          "payment",
		RideRequestID: req.RideRequestID, // Use this to identify the ride request in IPN to update transID
	}
	extraDataJSON, err := json.Marshal(extraData)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal extra data")
		return fmt.Errorf("failed to marshal extra data: %w", err)
	}
	extraDataBase64 := base64.StdEncoding.EncodeToString(extraDataJSON)

	// Build request signature
	var rawSignature bytes.Buffer
	rawSignature.WriteString("accessKey=")
	rawSignature.WriteString(p.cfg.MomoAccessKey)
	rawSignature.WriteString("&amount=")
	rawSignature.WriteString(fmt.Sprintf("%d", rideOffer.Fare)) // momo requires amount in integer
	rawSignature.WriteString("&extraData=")
	rawSignature.WriteString(extraDataBase64)
	rawSignature.WriteString("&orderId=")
	rawSignature.WriteString(requestID)
	rawSignature.WriteString("&orderInfo=")
	rawSignature.WriteString("Thanh toán chuyến đi")
	rawSignature.WriteString("&partnerClientId=")
	rawSignature.WriteString(userID.String())
	rawSignature.WriteString("&partnerCode=")
	rawSignature.WriteString(p.cfg.MomoPartnerCode)
	rawSignature.WriteString("&requestId=")
	rawSignature.WriteString(requestID)
	rawSignature.WriteString("&token=")
	rawSignature.WriteString(encryptedToken)

	// Sign request
	hmac := hmac.New(sha256.New, []byte(p.cfg.MomoSecretKey))
	hmac.Write(rawSignature.Bytes())
	signature := hex.EncodeToString(hmac.Sum(nil))

	// Build request payload
	payload := schemas.CheckoutRequest{
		PartnerClientID: userID.String(),
		PartnerCode:     p.cfg.MomoPartnerCode,
		RequestID:       requestID,
		Amount:          rideOffer.Fare,
		OrderID:         requestID,
		OrderInfo:       "Thanh toán chuyến đi",
		RedirectURL:     "",
		AutoCapture:     true,
		IpnURL:          p.cfg.MomoPaymentNotifyURL,
		ExtraData:       extraDataBase64,
		Token:           encryptedToken,
		Lang:            "vi",
		Signature:       signature,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal payload")
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Send request to MoMo API
	url := fmt.Sprintf("%s/%s", p.cfg.MomoPaymentURL, "tokenization/pay")
	log.Info().Str("url", url).Msg("Sending request to MoMo API")

	start := time.Now()
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Error().Err(err).Msg("Failed to send request to MoMo API")
		return fmt.Errorf("failed to send request to MoMo API: %w", err)
	}
	defer resp.Body.Close()

	duration := time.Since(start)
	log.Info().Dur("duration", duration).Int("statusCode", resp.StatusCode).Msg("Received response from MoMo API")

	// Read response from MoMo API
	var response schemas.CheckoutResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		log.Error().Err(err).Msg("Failed to decode response from MoMo API")
		return fmt.Errorf("failed to decode response from MoMo API: %w", err)
	}
	log.Debug().Interface("response", response).Msg("Decoded response from MoMo API")

	// Check if response is successful
	if response.ResultCode != 0 {
		log.Error().Int("resultCode", response.ResultCode).Str("message", response.Message).Msg("Checkout failed")
		return fmt.Errorf("checkout failed: %s", response.Message)
	}

	log.Info().Msg("Successfully completed CheckoutRide process")
	return nil
}

func (p *PaymentService) encryptRSA(tokenData schemas.TokenData) (string, error) {
	// Parse the PEM encoded public key
	block, _ := pem.Decode([]byte(p.cfg.MomoPublicKey))
	if block == nil {
		return "", fmt.Errorf("failed to parse PEM block containing the public key")
	}

	// Parse the public key
	pkixPub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse DER encoded public key: %w", err)
	}

	// Assert that the public key is an RSA key
	publicKey, ok := pkixPub.(*rsa.PublicKey)
	if !ok {
		return "", fmt.Errorf("not an RSA public key")
	}

	// Convert tokenData to JSON
	rawJsonData, err := json.Marshal(tokenData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal token data: %w", err)
	}

	// Encrypt the data
	ciphertext, err := rsa.EncryptPKCS1v15(
		rand.Reader,
		publicKey,
		rawJsonData,
	)
	if err != nil {
		return "", fmt.Errorf("encryption error: %w", err)
	}

	// Encode the encrypted data as base64
	hash := base64.StdEncoding.EncodeToString(ciphertext)

	return hash, nil
}

func (p *PaymentService) RefundRide(userID uuid.UUID, req schemas.RefundMomoRequest) error {
	log.Info().Msg("Starting RefundRide process")

	// Generate a new requestId
	requestID := uuid.New().String()

	// Get the ride request details
	rideRequest, err := p.repo.GetRideRequestByID(req.RideRequestID)
	if err != nil {
		log.Error().Err(err).Str("rideRequestID", req.RideRequestID.String()).Msg("Failed to get ride request details")
		return fmt.Errorf("failed to get ride request details: %w", err)
	}

	// Get ride offer details
	rideOffer, err := p.repo.GetRideOfferByID(req.RideOfferID)
	if err != nil {
		log.Error().Err(err).Str("rideOfferID", req.RideOfferID.String()).Msg("Failed to get ride offer details")
		return fmt.Errorf("failed to get ride offer details: %w", err)
	}

	// Build request signature
	var rawSignature bytes.Buffer
	rawSignature.WriteString("accessKey=")
	rawSignature.WriteString(p.cfg.MomoAccessKey)
	rawSignature.WriteString("&amount=")
	rawSignature.WriteString(fmt.Sprintf("%d", rideOffer.Fare))
	rawSignature.WriteString("&description=")
	rawSignature.WriteString("Hoàn tiền chuyến đi")
	rawSignature.WriteString("&orderId=")
	rawSignature.WriteString(requestID)
	rawSignature.WriteString("&partnerCode=")
	rawSignature.WriteString(p.cfg.MomoPartnerCode)
	rawSignature.WriteString("&requestId=")
	rawSignature.WriteString(requestID)
	rawSignature.WriteString("&transId=")
	rawSignature.WriteString(fmt.Sprintf("%d", rideRequest.MomoTransID))

	log.Debug().Str("rawSignature", rawSignature.String()).Msg("Built raw signature")

	// Sign request
	hmac := hmac.New(sha256.New, []byte(p.cfg.MomoSecretKey))
	hmac.Write(rawSignature.Bytes())
	signature := hex.EncodeToString(hmac.Sum(nil))

	log.Debug().Str("signature", signature).Msg("Generated signature")

	// Build request payload
	payload := schemas.MomoRefundRequest{
		PartnerCode: p.cfg.MomoPartnerCode,
		OrderID:     requestID,
		RequestID:   requestID,
		Amount:      rideOffer.Fare,
		TransID:     rideRequest.MomoTransID,
		Lang:        "vi",
		Description: "Hoàn tiền chuyến đi",
		Signature:   signature,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal payload")
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	log.Debug().RawJSON("payload", jsonPayload).Msg("Prepared request payload")

	// Send request to MoMo API
	url := fmt.Sprintf("%s/%s", p.cfg.MomoPaymentURL, "refund")
	log.Info().Str("url", url).Msg("Sending refund request to MoMo API")

	client := &http.Client{
		Timeout: time.Second * 30, // Set timeout to 30 seconds as per documentation
	}

	start := time.Now()
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Error().Err(err).Msg("Failed to send refund request to MoMo API")
		return fmt.Errorf("failed to send refund request to MoMo API: %w", err)
	}
	defer resp.Body.Close()

	duration := time.Since(start)
	log.Info().Dur("duration", duration).Int("statusCode", resp.StatusCode).Msg("Received response from MoMo API")

	// Read response from MoMo API
	var response schemas.MomoRefundResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		log.Error().Err(err).Msg("Failed to decode response from MoMo API")
		return fmt.Errorf("failed to decode response from MoMo API: %w", err)
	}

	log.Debug().Interface("response", response).Msg("Decoded response from MoMo API")

	// Check if response is successful
	if response.ResultCode != 0 {
		log.Error().Int("resultCode", response.ResultCode).Str("message", response.Message).Msg("Refund failed")
		return fmt.Errorf("refund failed: %s", response.Message)
	}

	// Update the refund status in your database
	// In fact only success ride match have transaction in db so no need to update refund status
	// err = p.repo.UpdateRefundStatus(req.RideID, "refunded", response.TransID)
	// if err != nil {
	// 	log.Error().Err(err).Str("rideID", req.RideID.String()).Msg("Failed to update refund status")
	// 	return fmt.Errorf("failed to update refund status: %w", err)
	// }

	log.Info().Msg("Successfully completed RefundRide process")
	return nil
}
