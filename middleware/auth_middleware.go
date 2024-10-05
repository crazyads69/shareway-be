package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"shareway/helper"
	"shareway/util/token"

	"github.com/gin-gonic/gin"
)

const (
	authorizationHeaderKey  = "authorization"
	authorizationTypeBearer = "bearer"
	authorizationPayloadKey = "authorization_payload"
)

var (
	ErrAuthHeaderMissing   = errors.New("authorization header is missing")
	ErrInvalidAuthFormat   = errors.New("invalid authorization header format")
	ErrUnsupportedAuthType = errors.New("unsupported authorization type")
)

// abortWithError is a helper function to abort the request with an error response
func abortWithError(ctx *gin.Context, status int, err error, messageEN, messageVI string) {
	response := helper.ErrorResponseWithMessage(err, messageEN, messageVI)
	ctx.AbortWithStatusJSON(status, response)
}

// AuthMiddleware creates a Gin middleware for authentication using PASETO tokens
func AuthMiddleware(maker token.PasetoMaker) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Extract the authorization header
		authHeader := ctx.GetHeader(authorizationHeaderKey)
		if authHeader == "" {
			abortWithError(ctx, http.StatusUnauthorized, ErrAuthHeaderMissing,
				"Authorization header is missing",
				"Thiếu header xác thực")
			return
		}

		// Split the header into fields
		fields := strings.Fields(authHeader)
		if len(fields) < 2 {
			abortWithError(ctx, http.StatusUnauthorized, ErrInvalidAuthFormat,
				"Invalid authorization header format",
				"Định dạng header xác thực không hợp lệ")
			return
		}

		// Check the authorization type
		authType := strings.ToLower(fields[0])
		if authType != authorizationTypeBearer {
			abortWithError(ctx, http.StatusUnauthorized,
				fmt.Errorf("%w: %s", ErrUnsupportedAuthType, authType),
				fmt.Sprintf("Unsupported authorization type: %s", authType),
				fmt.Sprintf("Loại xác thực không được hỗ trợ: %s", authType))
			return
		}

		// Extract the token
		tokenString := fields[1]

		// Verify the token
		payload, err := maker.VerifyToken(tokenString)
		if err != nil {
			abortWithError(ctx, http.StatusUnauthorized, err,
				"Invalid or expired token",
				"Token không hợp lệ hoặc đã hết hạn")
			return
		}

		// Set the verified payload in the context
		ctx.Set(authorizationPayloadKey, payload)
		ctx.Next()
	}
}
