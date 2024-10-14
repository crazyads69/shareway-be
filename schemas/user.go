package schemas

type GetUserProfileResponse struct {
	User UserResponse `json:"user" binding:"required"`
}
