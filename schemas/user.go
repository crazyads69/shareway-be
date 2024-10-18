package schemas

type GetUserProfileResponse struct {
	User UserResponse `json:"user" binding:"required"`
}

// Define RegisterDeviceTokenRequest struct
type RegisterDeviceTokenRequest struct {
	DeviceToken string `json:"device_token" binding:"required" validate:"required"`
}
