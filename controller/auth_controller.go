package controller

import (
	"time"

	"github.com/gin-gonic/gin"
)

type AuthController struct{}

type loginRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
}

type loginResponse struct {
	AccessToken           string    `json:"access_token"`
	RefreshToken          string    `json:"refresh_token"`
	AccessTokenExpiresAt  time.Time `json:"access_token_expires_at"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at"`
	// User                  userResponse `json:"user"`
}

func (ctrl *AuthController) Login(ctx *gin.Context) {
}
