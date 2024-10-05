package controller

import (
	"net/http"
	"shareride/helper"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
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
	log.Info().Msg("Login request received")

	var req loginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, helper.ErrorResponse(err))
		return
	}

	if req.Username != "admin" || req.Password != "admin123" {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"message": "Username or password is incorrect",
		})
		return
	}

	resp := loginResponse{
		AccessToken:           "1234",
		RefreshToken:          "1234",
		AccessTokenExpiresAt:  time.Now(),
		RefreshTokenExpiresAt: time.Now(),
	}

	ctx.JSON(http.StatusOK, resp)
}
