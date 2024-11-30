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

	"github.com/google/uuid"
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
	// Generate request ID for linking wallet
	requestID := uuid.New().String()

	// Store request ID to database
	err := p.repo.StoreRequestID(requestID, userID)
	if err != nil {
		return schemas.LinkWalletResponse{}, err
	}

	// Build request signature
	/* accessKey=$accessKey&amount=$amount&extraData=$extraData
	&ipnUrl=$ipnUrl&orderId=$orderId&orderInfo=$orderInfo
	&partnerClientId=$partnerClientId&partnerCode=
	$partnerCode&redirectUrl=$redirectUrl&requestId=
	$requestId&requestType=$requestType */

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
	rawSignature.WriteString(p.cfg.MomoPaymentNotifyURL)
	rawSignature.WriteString("&requestId=")
	rawSignature.WriteString(requestID)
	rawSignature.WriteString("&requestType=")
	rawSignature.WriteString("linkWallet")

	// Sign request
	// Create a new HMAC by defining the hash type and the key (as byte array)
	hmac := hmac.New(sha256.New, []byte(p.cfg.MomoSecretKey))
	hmac.Write(rawSignature.Bytes())

	// Get result and encode as hexadecimal string
	signature := hex.EncodeToString(hmac.Sum(nil))

	// Build request payload
	payload := schemas.LinkWalletRequest{
		PartnerCode:     p.cfg.MomoPartnerCode,
		AccessKey:       p.cfg.MomoAccessKey,
		RequestID:       requestID,
		Amount:          0,
		OrderID:         requestID, // Use request ID as order ID for now
		OrderInfo:       "Link wallet to user account",
		RedirectURL:     p.cfg.MomoPaymentNotifyURL,
		IpnURL:          p.cfg.MomoPaymentNotifyURL,
		PartnerClientID: userID.String(),
		ExtraData:       "",
		RequestType:     "linkWallet",
		Lang:            "vi",
		Signature:       signature,
	}

	// Send request to MoMo API
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return schemas.LinkWalletResponse{}, err
	}

	// Send request to MoMo API
	resp, err := http.Post(fmt.Sprintf("%s/%s", p.cfg.MomoPaymentURL, "create"), "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return schemas.LinkWalletResponse{}, err
	}

	// Read response from MoMo API
	var response schemas.LinkWalletResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return schemas.LinkWalletResponse{}, err
	}
	// Check if response is successful by checking result code
	if response.ResultCode != 0 {
		return schemas.LinkWalletResponse{}, fmt.Errorf("failed to link wallet: %s", response.Message)
	}

	// Log response
	fmt.Printf("Link wallet response: %v\n", response)
	return response, nil
}
