package schemas

type GetUserProfileResponse struct {
	User UserResponse `json:"user" binding:"required"`
}

// Define RegisterDeviceTokenRequest struct
type RegisterDeviceTokenRequest struct {
	DeviceToken string `json:"device_token" binding:"required" validate:"required"`
}

// Define RegisterDeviceTokenResponse struct
type UpdateUserProfileRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required,e164" validate:"required,e164"`
	FullName    string `json:"full_name" binding:"required,min=3,max=256" validate:"required,min=3,max=256"` // Email is optional
	Email       string `json:"email" binding:"omitempty,email,max=256" validate:"omitempty,email,max=256"`
}

type UpdateUserProfileResponse struct {
	User UserResponse `json:"user" binding:"required"`
}
