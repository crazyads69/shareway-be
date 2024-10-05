package helper

import (
	"shareway/schemas"

	"github.com/gin-gonic/gin"
)

// Response represents a standard API response structure
type Response struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	MessageEN string      `json:"message_en,omitempty"`
	MessageVI string      `json:"message_vi,omitempty"`
	Error     string      `json:"error,omitempty"`
}

// ErrorResponse returns a gin.H with error information
func ErrorResponse(err error) gin.H {
	return gin.H{
		"success": false,
		"error":   err.Error(),
	}
}

// ResponseWithMessage returns a Response with success status and messages in English and Vietnamese
func ResponseWithMessage(success bool, messageEN, messageVI string) Response {
	return Response{
		Success:   success,
		MessageEN: messageEN,
		MessageVI: messageVI,
	}
}

// SuccessResponse returns a Response with success status, data, and optional messages
func SuccessResponse(data interface{}, messageEN, messageVI string) Response {
	return Response{
		Success:   true,
		Data:      data,
		MessageEN: messageEN,
		MessageVI: messageVI,
	}
}

// ErrorResponseWithMessage returns a Response with error status and messages in English and Vietnamese
func ErrorResponseWithMessage(err error, messageEN, messageVI string) Response {
	return Response{
		Success:   false,
		Error:     err.Error(),
		MessageEN: messageEN,
		MessageVI: messageVI,
	}
}

// GinResponse is a helper function to send a JSON response using gin.Context
func GinResponse(c *gin.Context, statusCode int, response Response) {
	c.JSON(statusCode, response)
}

// ConvertToPayload attempts to convert an interface{} to a *schemas.Payload
func ConvertToPayload(data interface{}) (*schemas.Payload, bool) {
	payload, ok := data.(*schemas.Payload)
	return payload, ok
}
