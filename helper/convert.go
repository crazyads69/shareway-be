package helper

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"

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
	// Hash the UUID for good distribution
	hasher := sha256.New()
	hasher.Write(id[:])
	hash := hasher.Sum(nil)

	// Convert first 4 bytes to uint32
	uid := binary.BigEndian.Uint32(hash[:4])

	// Ensure the UID is between 1 and (2^32 - 1)
	if uid == 0 {
		return 1
	}

	// Return the UUID as a uint32
	return uid
}
