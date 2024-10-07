package fpt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"shareway/util"
)

// FPTReader struct holds the configuration for FPT AI API
type FPTReader struct {
	cfg util.Config
}

// CCCDInfo holds the information extracted from a CCCD image
type CCCDInfo struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	DOB         string `json:"dob,omitempty"`
	Sex         string `json:"sex,omitempty"`
	Nationality string `json:"nationality,omitempty"`
	Home        string `json:"home,omitempty"`
	Address     string `json:"address,omitempty"`
	DOE         string `json:"doe,omitempty"`
	Features    string `json:"features,omitempty"`
	IssueDate   string `json:"issue_date,omitempty"`
	Type        string `json:"type,omitempty"`
}

// NewFPTReader creates a new instance of FPTReader
func NewFPTReader(cfg util.Config) *FPTReader {
	return &FPTReader{
		cfg: cfg,
	}
}

// mapToStruct maps a map[string]interface{} to a struct
func mapToStruct(m map[string]interface{}, s interface{}) error {
	jsonData, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonData, s)
}

// VerifyImageWithFPTAI sends an image to FPT AI for verification and returns the extracted information
func (r *FPTReader) VerifyImageWithFPTAI(image *multipart.FileHeader) (*CCCDInfo, error) {
	// Open the image file
	src, err := image.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open image: %w", err)
	}
	defer src.Close()

	// Prepare the multipart form data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("image", image.Filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	// Copy the image data to the form file
	if _, err = io.Copy(part, src); err != nil {
		return nil, fmt.Errorf("failed to copy image data: %w", err)
	}

	// Close the multipart writer
	if err = writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Create and send the HTTP request
	req, err := http.NewRequest("POST", r.cfg.FptAiApiUrl, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("api-key", r.cfg.FptAiApiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read and parse the response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	// Check for errors in the response
	if errorCode, ok := result["errorCode"].(float64); ok && errorCode != 0 {
		errorMessage, _ := result["errorMessage"].(string)
		return nil, fmt.Errorf("API error: code %v, message: %s", errorCode, errorMessage)
	}

	// Extract the information from the response
	data, ok := result["data"].([]interface{})
	if !ok || len(data) == 0 {
		return nil, fmt.Errorf("invalid or empty data in response")
	}

	firstItem, ok := data[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data format in response")
	}

	info := &CCCDInfo{}
	if err := mapToStruct(firstItem, info); err != nil {
		return nil, fmt.Errorf("failed to extract CCCD info: %w", err)
	}

	return info, nil
}
