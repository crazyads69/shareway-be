package schemas

type MoMoIPN struct {
	PartnerCode     string `json:"partnerCode"`
	OrderID         string `json:"orderId"`
	RequestID       string `json:"requestId"`
	Amount          int64  `json:"amount"`
	OrderInfo       string `json:"orderInfo"`
	OrderType       string `json:"orderType"`
	PartnerClientID string `json:"partnerClientId"`
	CallbackToken   string `json:"callbackToken"`
	TransID         int64  `json:"transId"`
	ResultCode      int    `json:"resultCode"`
	Message         string `json:"message"`
	PayType         string `json:"payType"`
	ResponseTime    int64  `json:"responseTime"`
	ExtraData       string `json:"extraData"`
	Signature       string `json:"signature"`
}

type TokenizationBindRequest struct {
	PartnerCode     string `json:"partnerCode"`
	CallbackToken   string `json:"callbackToken"`
	RequestID       string `json:"requestId"`
	OrderID         string `json:"orderId"`
	PartnerClientID string `json:"partnerClientId"`
	Lang            string `json:"lang"`
	Signature       string `json:"signature"`
}

type TokenizationBindResponse struct {
	PartnerCode     string `json:"partnerCode"`
	RequestID       string `json:"requestId"`
	OrderID         string `json:"orderId"`
	AESToken        string `json:"aesToken"`
	ResultCode      int    `json:"resultCode"`
	PartnerClientID string `json:"partnerClientId"`
	ResponseTime    int64  `json:"responseTime"`
	Message         string `json:"message"`
}

type DecodedToken struct {
	Value     string `json:"value"`
	UserAlias string `json:"userAlias"`
	ProfileID string `json:"profileId"`
}
