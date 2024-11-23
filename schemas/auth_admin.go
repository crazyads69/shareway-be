package schemas

import (
	"time"

	"github.com/google/uuid"
)

// Define LoginAdminRequest struct
type LoginAdminRequest struct {
	Username string `json:"username" binding:"required" validate:"required,min=1,max=255"`
	Password string `json:"password" binding:"required" validate:"required,min=6,max=255"`
}

type AdminInfo struct {
	ID        uuid.UUID `json:"admin_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Username  string    `json:"username"`
	FullName  string    `json:"full_name"`
	Role      string    `json:"role"`
}

// Define LoginAdminResponse struct
type LoginAdminResponse struct {
	Token     string    `json:"token"`
	AdminInfo AdminInfo `json:"admin_info"`
}
