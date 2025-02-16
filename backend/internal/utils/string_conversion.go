package utils

import (
	"strings"
	"unicode"
)

func RemoveNonAlphanumericPrefix(input string) string {
	return strings.TrimFunc(input, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
}

// ToCamelCase
// Function to convert a string to CamelCase (handles alphanumeric input)
func ToCamelCase(input string) string {
	var result []rune
	capitalizeNext := false

	input = RemoveNonAlphanumericPrefix(input)

	for i, r := range input {
		if i == 0 {
			// Always lowercase the first character
			result = append(result, unicode.ToLower(r))
			continue
		}

		if unicode.IsSpace(r) || r == '_' || r == '-' || !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			// Set the flag to capitalize the next valid character
			capitalizeNext = true
			continue
		}

		if capitalizeNext {
			// Capitalize the current character then reset the flag
			result = append(result, unicode.ToUpper(r))
			capitalizeNext = false
		} else {
			// Append as it is
			result = append(result, r)
		}
	}

	if len(result) == 0 {
		return ""
	}

	return string(result)
}
