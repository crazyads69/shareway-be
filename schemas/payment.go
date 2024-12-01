package schemas

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
	Type string `json:"type"`
	// Thêm các trường khác nếu cần
}
