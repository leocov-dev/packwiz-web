package utils

import (
	"testing"
	"unicode"
)

func TestGenerateLinkToken(t *testing.T) {
	t.Run("generates string of correct length", func(t *testing.T) {
		lengths := []int{4, 5, 10, 32}
		for _, length := range lengths {
			result := GenerateLinkToken(length)
			if len(result) != length {
				t.Errorf("Expected length %d, got %d", length, len(result))
			}
		}
	})

	t.Run("contains only alphanumeric characters", func(t *testing.T) {
		result := GenerateLinkToken(100) // Using a longer string to test character set
		for i, char := range result {
			if !unicode.IsLetter(char) && !unicode.IsNumber(char) {
				t.Errorf("Invalid character at position %d: %c", i, char)
			}
		}
	})

	t.Run("generates different strings", func(t *testing.T) {
		length := 10
		iterations := 5
		seen := make(map[string]bool)

		for i := 0; i < iterations; i++ {
			result := GenerateLinkToken(length)
			if seen[result] {
				t.Error("Generated duplicate string")
			}
			seen[result] = true
		}
	})
}
