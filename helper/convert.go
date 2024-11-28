package helper

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"shareway/util/sanctum"

	"github.com/google/uuid"
)

// ConvertToStringMap converts an interface{} to map[string]string
// It maintains the original structure by JSON marshaling nested objects
func ConvertToStringMap(data interface{}) (map[string]string, error) {
	result := make(map[string]string)

	// Handle nil input
	if data == nil {
		return result, nil
	}

	// First convert the data to a map[string]interface{} using JSON marshal/unmarshal
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("error marshaling data: %v", err)
	}

	var intermediate map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &intermediate); err != nil {
		return nil, fmt.Errorf("error unmarshaling data: %v", err)
	}

	// Convert the intermediate map to map[string]string
	for key, value := range intermediate {
		switch v := value.(type) {
		case string:
			result[key] = v
		case float64:
			if float64(int64(v)) == v {
				// It's an integer
				result[key] = fmt.Sprintf("%d", int64(v))
			} else {
				result[key] = fmt.Sprintf("%.2f", v)
			}
		case bool:
			result[key] = fmt.Sprintf("%v", v)
		case nil:
			result[key] = ""
		default:
			// For complex objects (maps, arrays, nested structs), keep them as JSON strings
			nestedJSON, err := json.Marshal(v)
			if err != nil {
				return nil, fmt.Errorf("error marshaling nested object for key %s: %v", key, err)
			}
			result[key] = string(nestedJSON)
		}
	}

	return result, nil
}

// Recommended method: Combines hash and range checking
func UuidToUid(id uuid.UUID) uint32 {
	// Agora requires uid to be between 1 and (2^32 - 1)
	// We should ensure consistent mapping from UUID to uint32

	// Use last 4 bytes of UUID instead of hashing
	// This ensures the same UUID always maps to the same UID
	bytes := id[12:16]
	uid := binary.BigEndian.Uint32(bytes)

	// Ensure uid is not 0 (reserved in Agora)
	if uid == 0 {
		uid = 1
	}

	return uid
}

// ConvertToAdminPayload converts the payload to a map of string and interface
func ConvertToAdminPayload(data interface{}) (*sanctum.SanctumTokenPayload, error) {
	payload, ok := data.(*sanctum.SanctumTokenPayload)
	if !ok {
		return nil, fmt.Errorf("failed to convert payload")
	}
	return payload, nil
}
