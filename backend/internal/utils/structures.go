package utils

import "strings"

var sensitiveKeyMarkers = []string{"password", "secret", "token"}

func isSensitiveKey(key string) bool {
	lower := strings.ToLower(key)
	for _, marker := range sensitiveKeyMarkers {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

// RedactSensitiveValues returns a copy of src with any value whose key looks
// like it holds a credential (password, secret, token, ...) replaced with a
// redaction placeholder.
func RedactSensitiveValues(src map[string]interface{}) map[string]interface{} {
	dst := make(map[string]interface{}, len(src))
	for key, value := range src {
		if isSensitiveKey(key) {
			dst[key] = "********"
			continue
		}
		dst[key] = value
	}
	return dst
}

func DeepCopyMapStringSlice(src map[string][]string) map[string][]string {
	dst := make(map[string][]string)

	for key, slice := range src {
		copiedSlice := make([]string, len(slice))
		copy(copiedSlice, slice)

		dst[key] = copiedSlice
	}

	return dst
}
