package service

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"shareway/infra/ws"
	"shareway/repository"
	"shareway/schemas"
	"shareway/util"
	"strconv"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type IPNService struct {
	repo repository.IIPNRepository
	hub  *ws.Hub
	cfg  util.Config
}

type IIPNService interface {
	VerifyIPN(schemas.MoMoIPN) bool
	HandleLinkWalletCallback(schemas.MoMoIPN) error
	DecryptAESToken(string) (schemas.DecodedToken, error)
	HandleIPN(schemas.MoMoIPN) error
}

func NewIPNService(repo repository.IIPNRepository, hub *ws.Hub, cfg util.Config) IIPNService {
	return &IPNService{
		repo: repo,
		hub:  hub,
		cfg:  cfg,
	}
}

func (s *IPNService) VerifyIPN(ipn schemas.MoMoIPN) bool {
	// Verify IPN signature
	/* accessKey=$accessKey&amount=$amount&callbackToken=
	$callbackToken&extraData=$extraData&message=$message
	&orderId=$orderId&orderInfo=$orderInfo&orderType=
	$orderType&partnerClientId=$partnerClientId
	&partnerCode=$partnerCode&payType=$payType&requestId=
	$requestId&responseTime=$responseTime&resultCode=
	$resultCode&transId=$transId */

	var rawSignature bytes.Buffer
	rawSignature.WriteString("accessKey=")
	rawSignature.WriteString(s.cfg.MomoAccessKey)
	rawSignature.WriteString("&amount=")
	rawSignature.WriteString(strconv.FormatInt(ipn.Amount, 10))
	rawSignature.WriteString("&callbackToken=")
	rawSignature.WriteString(ipn.CallbackToken)
	rawSignature.WriteString("&extraData=")
	rawSignature.WriteString(ipn.ExtraData)
	rawSignature.WriteString("&message=")
	rawSignature.WriteString(ipn.Message)
	rawSignature.WriteString("&orderId=")
	rawSignature.WriteString(ipn.OrderID)
	rawSignature.WriteString("&orderInfo=")
	rawSignature.WriteString(ipn.OrderInfo)
	rawSignature.WriteString("&orderType=")
	rawSignature.WriteString(ipn.OrderType)
	rawSignature.WriteString("&partnerClientId=")
	rawSignature.WriteString(ipn.PartnerClientID)
	rawSignature.WriteString("&partnerCode=")
	rawSignature.WriteString(ipn.PartnerCode)
	rawSignature.WriteString("&payType=")
	rawSignature.WriteString(ipn.PayType)
	rawSignature.WriteString("&requestId=")
	rawSignature.WriteString(ipn.RequestID)
	rawSignature.WriteString("&responseTime=")
	rawSignature.WriteString(strconv.FormatInt(ipn.ResponseTime, 10))
	rawSignature.WriteString("&resultCode=")
	rawSignature.WriteString(strconv.Itoa(ipn.ResultCode))
	rawSignature.WriteString("&transId=")
	rawSignature.WriteString(strconv.FormatInt(ipn.TransID, 10))
	h := hmac.New(sha256.New, []byte(s.cfg.MomoSecretKey))
	h.Write([]byte(rawSignature.Bytes()))
	signature := hex.EncodeToString(h.Sum(nil))
	return signature == ipn.Signature
}

func (s *IPNService) HandleLinkWalletCallback(ipn schemas.MoMoIPN) error {
	log.Info().
		Str("partnerClientID", ipn.PartnerClientID).
		Str("callbackToken", ipn.CallbackToken).
		Msg("Starting HandleLinkWalletCallback")

	// Get user from partner client ID
	user, err := s.repo.GetUserByPartnerClientID(ipn.PartnerClientID)
	if err != nil {
		log.Error().Err(err).Str("partnerClientID", ipn.PartnerClientID).Msg("Failed to get user")
		return fmt.Errorf("failed to get user: %w", err)
	}
	log.Info().Str("userID", user.ID.String()).Msg("Retrieved user")

	// Store callback token
	err = s.repo.StoreCallbackToken(ipn.CallbackToken, user.ID)
	if err != nil {
		log.Error().Err(err).Str("userID", user.ID.String()).Msg("Failed to store callback token")
		return fmt.Errorf("failed to store callback token: %w", err)
	}
	log.Info().Str("userID", user.ID.String()).Msg("Stored callback token")

	// Create signature
	requestID := uuid.New().String()
	var rawSignature bytes.Buffer
	rawSignature.WriteString("accessKey=")
	rawSignature.WriteString(s.cfg.MomoAccessKey)
	rawSignature.WriteString("&callbackToken=")
	rawSignature.WriteString(ipn.CallbackToken)
	rawSignature.WriteString("&orderId=")
	rawSignature.WriteString(user.MomoFirstRequestID.String())
	rawSignature.WriteString("&partnerClientId=")
	rawSignature.WriteString(ipn.PartnerClientID)
	rawSignature.WriteString("&partnerCode=")
	rawSignature.WriteString(s.cfg.MomoPartnerCode)
	rawSignature.WriteString("&requestId=")
	rawSignature.WriteString(requestID)

	// Sign request
	h := hmac.New(sha256.New, []byte(s.cfg.MomoSecretKey))
	h.Write(rawSignature.Bytes())
	signature := hex.EncodeToString(h.Sum(nil))
	log.Debug().Str("signature", signature).Msg("Created signature for tokenization bind request")

	// Build RecurringToken request
	payload := schemas.TokenizationBindRequest{
		PartnerCode:     s.cfg.MomoPartnerCode,
		CallbackToken:   ipn.CallbackToken,
		RequestID:       requestID,
		OrderID:         user.MomoFirstRequestID.String(),
		PartnerClientID: ipn.PartnerClientID,
		Lang:            "vi",
		Signature:       signature,
	}

	// Send request to MoMo API
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal payload")
		return fmt.Errorf("failed to marshal payload: %w", err)
	}
	log.Debug().RawJSON("payload", jsonPayload).Msg("Prepared request payload")

	url := fmt.Sprintf("%s/%s", s.cfg.MomoPaymentURL, "tokenization/bind")
	log.Info().Str("url", url).Msg("Sending request to MoMo API")
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Error().Err(err).Str("url", url).Msg("Failed to send request to MoMo API")
		return fmt.Errorf("failed to send request to MoMo API: %w", err)
	}
	defer resp.Body.Close()

	// Decode response
	var response schemas.TokenizationBindResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		log.Error().Err(err).Msg("Failed to decode response")
		return fmt.Errorf("failed to decode response: %w", err)
	}
	log.Debug().Interface("response", response).Msg("Received response from MoMo API")

	// Check status code
	if response.ResultCode != 0 {
		log.Error().Int("resultCode", response.ResultCode).Str("message", response.Message).Msg("MoMo API error")
		return fmt.Errorf("MoMo API error: %s", response.Message)
	}

	// Decrypt AES token
	decodedToken, err := s.DecryptAESToken(response.AESToken)
	if err != nil {
		log.Error().Err(err).Msg("Failed to decrypt AES token")
		return fmt.Errorf("failed to decrypt AES token: %w", err)
	}
	log.Info().Str("userAlias", decodedToken.UserAlias).Msg("Successfully decrypted AES token")

	// Update user with MoMo token
	err = s.repo.UpdateUserMoMoToken(user.ID, decodedToken)
	if err != nil {
		log.Error().Err(err).Str("userID", user.ID.String()).Msg("Failed to update user MoMo token")
		return fmt.Errorf("failed to update user MoMo token: %w", err)
	}
	log.Info().Str("userID", user.ID.String()).Msg("Updated user with MoMo token")

	log.Info().Msg("Successfully completed HandleLinkWalletCallback")
	return nil
}
func (s *IPNService) DecryptAESToken(encryptedToken string) (schemas.DecodedToken, error) {
	key := []byte(s.cfg.MomoSecretKey)
	ciphertext, _ := base64.StdEncoding.DecodeString(encryptedToken)

	block, err := aes.NewCipher(key)
	if err != nil {
		return schemas.DecodedToken{}, err
	}

	if len(ciphertext) < aes.BlockSize {
		return schemas.DecodedToken{}, fmt.Errorf("ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(ciphertext, ciphertext)

	// Xóa padding
	padding := int(ciphertext[len(ciphertext)-1])
	ciphertext = ciphertext[:len(ciphertext)-padding]

	var decodedToken schemas.DecodedToken
	err = json.Unmarshal(ciphertext, &decodedToken)
	if err != nil {
		return schemas.DecodedToken{}, err
	}

	return decodedToken, nil
}

func (s *IPNService) HandleIPN(ipn schemas.MoMoIPN) error {
	return nil
}
