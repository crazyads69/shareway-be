package schemas

import "mime/multipart"

type GetUserProfileResponse struct {
	User UserResponse `json:"user" binding:"required"`
}

// Define RegisterDeviceTokenRequest struct
type RegisterDeviceTokenRequest struct {
	DeviceToken string `json:"device_token" binding:"required" validate:"required"`
}

// Define RegisterDeviceTokenResponse struct
type UpdateUserProfileRequest struct {
	FullName string `json:"full_name" binding:"required,min=3,max=256" validate:"required,min=3,max=256"`
	Email    string `json:"email" binding:"omitempty,email,max=256" validate:"omitempty,email,max=256"`
	Gender   string `json:"gender" binding:"required,oneof=male female" validate:"required,oneof=male female"`
}

type UpdateUserProfileResponse struct {
	User UserResponse `json:"user" binding:"required"`
}

// Define UpdateAvatarRequest struct
type UpdateAvatarRequest struct {
	AvatarImage *multipart.FileHeader `form:"avatar_image" binding:"required" validate:"required"`
}

// Define UpdateAvatarResponse struct
type UpdateAvatarResponse struct {
	User UserResponse `json:"user" binding:"required"`
}
