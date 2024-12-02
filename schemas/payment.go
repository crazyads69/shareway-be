package schemas

import "github.com/google/uuid"

type LinkMomoRequest struct {
	// Different from phoneNumber that is used for login
	// This is the phone number that is registered with momo wallet
	WalletPhoneNumber string `json:"walletPhoneNumber" binding:"required" validate:"required,min=10,max=10"`
}

type LinkWalletRequest struct {
	PartnerCode     string `json:"partnerCode"`
	AccessKey       string `json:"accessKey"`
	RequestID       string `json:"requestId"`
	Amount          int64  `json:"amount"`
	OrderID         string `json:"orderId"`
	OrderInfo       string `json:"orderInfo"`
	RedirectURL     string `json:"redirectUrl"`
	IpnURL          string `json:"ipnUrl"`
	PartnerClientID string `json:"partnerClientId"`
	ExtraData       string `json:"extraData"`
	RequestType     string `json:"requestType"`
	Lang            string `json:"lang"`
	Signature       string `json:"signature"`
}

type LinkWalletResponse struct {
	PartnerCode     string `json:"partnerCode"`
	RequestID       string `json:"requestId"`
	OrderID         string `json:"orderId"`
	Amount          int64  `json:"amount"`
	ResponseTime    int64  `json:"responseTime"`
	Message         string `json:"message"`
	ResultCode      int    `json:"resultCode"`
	PayUrl          string `json:"payUrl"`
	QrCodeUrl       string `json:"qrCodeUrl"`
	Deeplink        string `json:"deeplink"`
	DeeplinkMiniApp string `json:"deeplinkMiniApp"`
	PartnerClientID string `json:"partnerClientId"`
}

type LinkMomoWalletResponse struct {
	Deeplink string `json:"deeplink"` // send this to fe flutter app for open momo and perform linked
}

type ExtraData struct {
	Type          string    `json:"type"`
	RideRequestID uuid.UUID `json:"rideRequestID"`
	// Thêm các trường khác nếu cần
}

type CheckoutRequest struct {
	StoreID         string `json:"storeId"`
	RequestID       string `json:"requestId"`
	OrderID         string `json:"orderId"`
	RedirectURL     string `json:"redirectUrl"`
	IpnURL          string `json:"ipnUrl"`
	PartnerClientID string `json:"partnerClientId"`
	PartnerCode     string `json:"partnerCode"`
	PartnerName     string `json:"partnerName"`
	Amount          int64  `json:"amount"`
	OrderInfo       string `json:"orderInfo"`
	ExtraData       string `json:"extraData"`
	AutoCapture     bool   `json:"autoCapture"`
	Lang            string `json:"lang"`
	Token           string `json:"token"`
	Signature       string `json:"signature"`
}
type CheckoutResponse struct {
	PartnerCode     string `json:"partnerCode"`
	OrderID         string `json:"orderId"`
	RequestID       string `json:"requestId"`
	Amount          int64  `json:"amount"`
	ResponseTime    int64  `json:"responseTime"`
	Message         string `json:"message"`
	ResultCode      int    `json:"resultCode"`
	PayURL          string `json:"payUrl"`
	Deeplink        string `json:"deeplink"`
	QRCodeURL       string `json:"qrCodeUrl"`
	TransID         int64  `json:"transId"`
	PartnerClientID string `json:"partnerClientId"`
}

type TokenData struct {
	Value               string `json:"value"`
	RequireSecurityCode bool   `json:"requireSecurityCode"`
}

type CheckoutRideRequest struct {
	// Use for checkout with momo
	// The ID of the ride request (current user is the hitcher)
	RideRequestID uuid.UUID `json:"rideRequestID" binding:"required,uuid" validate:"required,uuid"`
	// The ID of the ride offer (the user who received request is the driver)
	RideOfferID uuid.UUID `json:"rideOfferID" binding:"required,uuid" validate:"required,uuid"`
	// The ID of the receiver (the user who received the request) aka the driver
	ReceiverID uuid.UUID `json:"receiverID" binding:"required,uuid" validate:"required,uuid"`
}

// type CheckoutRideResponse struct {
// }

// Define RefundMomoRequest
type RefundMomoRequest struct {
	// The ID of the ride request (current user is the hitcher)
	// The ride request contains then transaction ID from momo so could use for refund when needed (ride canceled, cannot create ride, etc)
	RideRequestID uuid.UUID `json:"rideRequestID" binding:"required,uuid" validate:"required,uuid"`
	RideOfferID   uuid.UUID `json:"rideOfferID" binding:"required,uuid" validate:"required,uuid"`
}

type MomoRefundRequest struct {
	PartnerCode string `json:"partnerCode"`
	OrderID     string `json:"orderId"`
	RequestID   string `json:"requestId"`
	Amount      int64  `json:"amount"`
	TransID     int64  `json:"transId"`
	Lang        string `json:"lang"`
	Description string `json:"description"`
	Signature   string `json:"signature"`
}

type MomoRefundResponse struct {
	PartnerCode  string `json:"partnerCode"`
	OrderID      string `json:"orderId"`
	RequestID    string `json:"requestId"`
	Amount       int64  `json:"amount"`
	TransID      int64  `json:"transId"`
	ResultCode   int    `json:"resultCode"`
	Message      string `json:"message"`
	ResponseTime int64  `json:"responseTime"`
}
