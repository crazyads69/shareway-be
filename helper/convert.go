package helper

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/google/uuid"
)

// StructToFCMData converts any struct to a map[string]string suitable for FCM data payload
func StructToFCMData(obj interface{}) map[string]string {
	return structToFCMDataRecursive(reflect.ValueOf(obj), "")
}

func structToFCMDataRecursive(v reflect.Value, prefix string) map[string]string {
	result := make(map[string]string)

	// Handle nil values
	if !v.IsValid() {
		return result
	}

	// Dereference pointers
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return result
		}
		v = v.Elem()
	}

	// Check for empty prefix for primitive types
	if prefix == "" {
		prefix = "value" // or another suitable default key
	}

	switch v.Kind() {
	case reflect.Struct:
		if v.Type() == reflect.TypeOf(time.Time{}) {
			result[prefix] = v.Interface().(time.Time).Format(time.RFC3339)
			return result
		}
		if v.Type() == reflect.TypeOf(uuid.UUID{}) {
			result[prefix] = v.Interface().(uuid.UUID).String()
			return result
		}
		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			fieldType := v.Type().Field(i)

			// Skip unexported fields
			if !field.CanInterface() {
				continue
			}

			key := fieldType.Tag.Get("json")
			if key == "" {
				key = fieldType.Name
			} else {
				// Remove any additional json tag options (like omitempty)
				if comma := strings.Index(key, ","); comma != -1 {
					key = key[:comma]
				}
			}

			if prefix != "" {
				key = prefix + "_" + key
			}

			nestedResult := structToFCMDataRecursive(field, key)
			for k, v := range nestedResult {
				result[k] = v
			}
		}
	case reflect.Map:
		for _, key := range v.MapKeys() {
			keyStr := fmt.Sprintf("%v", key.Interface())
			if prefix != "" {
				keyStr = prefix + "_" + keyStr
			}
			nestedResult := structToFCMDataRecursive(v.MapIndex(key), keyStr)
			for k, v := range nestedResult {
				result[k] = v
			}
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			key := fmt.Sprintf("%s_%d", prefix, i)
			nestedResult := structToFCMDataRecursive(v.Index(i), key)
			for k, v := range nestedResult {
				result[k] = v
			}
		}
	case reflect.String:
		result[prefix] = v.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		result[prefix] = fmt.Sprintf("%d", v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		result[prefix] = fmt.Sprintf("%d", v.Uint())
	case reflect.Float32, reflect.Float64:
		result[prefix] = fmt.Sprintf("%.2f", v.Float())
	case reflect.Bool:
		result[prefix] = fmt.Sprintf("%t", v.Bool())
	default:
		// For any other types, convert to string
		result[prefix] = fmt.Sprintf("%v", v.Interface())
	}

	return result
}
