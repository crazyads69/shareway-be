package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"shareride/helper"

	"github.com/gin-gonic/gin"
)

const (
	AuthorizationHeaderKey  = "authorization"
	AuthorizationTypeBearer = "bearer"
	AuthorizationPayloadKey = "authorization_payload"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authorizationHeader := ctx.GetHeader(AuthorizationHeaderKey)
		if len(authorizationHeader) == 0 {
			err := errors.New("authorization is not provided")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, helper.ErrorResponse(err))
			return
		}

		fields := strings.Fields(authorizationHeader)
		if len(fields) < 2 {
			err := errors.New("invalid authorization header format")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, helper.ErrorResponse(err))
			return
		}

		authorizationHeaderType := strings.ToLower(fields[0])
		if authorizationHeaderType != AuthorizationTypeBearer {
			err := fmt.Errorf("unsupported authorization type %s", authorizationHeaderType)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, helper.ErrorResponse(err))
			return
		}

		token := fields[1]
		// Do some verification logic to get payload
		// payload, err := tokenMaker.VerifyToken(token)
		// if err != nil {
		// 	ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
		// 	return
		// }

		ctx.Set(AuthorizationPayloadKey, token) // TODO: replace with payload
		ctx.Next()
	}
}
