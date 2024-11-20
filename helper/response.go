package helper

import (
	"fmt"
	"shareway/infra/fpt"
	"shareway/schemas"
	"time"

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
func ConvertToPayload(data interface{}) (*schemas.Payload, error) {
	payload, ok := data.(*schemas.Payload)
	if !ok {
		return nil, fmt.Errorf("failed to convert payload")
	}
	return payload, nil
}

// ValidateCCCDInfo checks the validity of CCCD (Citizen Identity Card) information
// from both front and back sides of the card.
func ValidateCCCDInfo(frontCCCDInfo, backCCCDInfo *fpt.CCCDInfo) error {
	// Define constants
	const (
		dateLayout = "02/01/2006" // Adjust this if your date format is different
		minAge     = 18
	)

	// Parse dates
	dates, err := parseDates(dateLayout, frontCCCDInfo.DOE, backCCCDInfo.IssueDate, frontCCCDInfo.DOB)
	if err != nil {
		return err
	}

	doe, issueDate, dob := dates[0], dates[1], dates[2]
	currentDate := time.Now().UTC()

	// Validate expiry date (DOE)
	if err := validateExpiry(doe, currentDate, issueDate); err != nil {
		return err
	}

	// Validate date of birth (DOB)
	if err := validateDOB(dob, currentDate, issueDate, minAge); err != nil {
		return err
	}

	return nil
}

// parseDates parses multiple date strings into time.Time objects
func parseDates(layout string, dates ...string) ([]time.Time, error) {
	parsedDates := make([]time.Time, len(dates))
	for i, date := range dates {
		parsed, err := time.Parse(layout, date)
		if err != nil {
			return nil, fmt.Errorf("invalid date format for %s: %w", date, err)
		}
		parsedDates[i] = parsed
	}
	return parsedDates, nil
}

// validateExpiry checks if the DOE is valid
func validateExpiry(doe, currentDate, issueDate time.Time) error {
	if doe.Before(currentDate) {
		return fmt.Errorf("CCCD has expired: expiry date %v is before current date %v", doe, currentDate)
	}
	if doe.Before(issueDate) || doe.Equal(issueDate) {
		return fmt.Errorf("invalid DOE: expiry date %v is not after issue date %v", doe, issueDate)
	}
	return nil
}

// validateDOB checks if the DOB is valid and if the person is of legal age
func validateDOB(dob, currentDate, issueDate time.Time, minAge int) error {
	if dob.After(currentDate) {
		return fmt.Errorf("invalid DOB: date of birth %v is in the future", dob)
	}
	if dob.After(issueDate) {
		return fmt.Errorf("invalid DOB: date of birth %v is after issue date %v", dob, issueDate)
	}
	age := currentDate.Year() - dob.Year()
	if currentDate.YearDay() < dob.YearDay() {
		age--
	}
	if age < minAge {
		return fmt.Errorf("user is under %d years old: age %d is less than the minimum age", minAge, age)
	}
	return nil
}


// Helper function to validate image types
func IsValidImageType(contentType string) bool {
    validTypes := map[string]bool{
        "image/jpeg": true,
        "image/png":  true,
        "image/gif":  true,
        "image/webp": true,
		"image/jpg":  true,
    }
    return validTypes[contentType]
}
