package utils

import (
	"strings"
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

func TestGeneratePassword(t *testing.T) {
	t.Run("length and minimum", func(t *testing.T) {
		if got := len(GeneratePassword(20)); got != 20 {
			t.Errorf("expected 20, got %d", got)
		}
		if got := len(GeneratePassword(3)); got != 12 {
			t.Errorf("expected minimum 12, got %d", got)
		}
	})

	t.Run("has letter and digit, only allowed chars", func(t *testing.T) {
		for i := 0; i < 500; i++ {
			pw := GeneratePassword(12)
			hasLetter, hasDigit := false, false
			for _, ch := range pw {
				switch {
				case ch >= 'a' && ch <= 'z', ch >= 'A' && ch <= 'Z':
					hasLetter = true
				case ch >= '0' && ch <= '9':
					hasDigit = true
				case !strings.ContainsRune(passwordSymbols, ch):
					t.Fatalf("invalid character %q in %q", ch, pw)
				}
			}
			if !hasLetter || !hasDigit {
				t.Fatalf("password %q missing letter or digit", pw)
			}
		}
	})

	t.Run("generates different passwords", func(t *testing.T) {
		if GeneratePassword(20) == GeneratePassword(20) {
			t.Error("expected different passwords")
		}
	})
}
