package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"shareway/util/sanctum"

	"github.com/gin-gonic/gin"
)

func AuthAdminMiddleware(sanctumToken *sanctum.SanctumToken) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Extract the authorization header
		authHeader := ctx.GetHeader(AuthorizationHeaderKey)
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
		if authType != AuthorizationTypeBearer {
			abortWithError(ctx, http.StatusUnauthorized,
				fmt.Errorf("%w: %s", ErrUnsupportedAuthType, authType),
				fmt.Sprintf("Unsupported authorization type: %s", authType),
				fmt.Sprintf("Loại xác thực không được hỗ trợ: %s", authType))
			return
		}

		// Extract the token
		tokenString := fields[1]

		// Verify the token
		payload, err := sanctumToken.VerifySanctumToken(tokenString)
		if err != nil {
			abortWithError(ctx, http.StatusUnauthorized, err,
				"Invalid token",
				"Token không hợp lệ")
			return
		}

		// Set the verified payload in the context
		ctx.Set(AuthorizationPayloadKey, payload)
		ctx.Next()
	}
}
