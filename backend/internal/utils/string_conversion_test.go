package utils

import (
	"testing"
)

func TestToCamelCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Basic cases
		{"hello_world", "helloWorld"},
		{"convert-to-camel", "convertToCamel"},
		{"make it camel case", "makeItCamelCase"},
		{"AlreadyCamelCase", "alreadyCamelCase"},

		// Special characters and numbers
		{"123_test case", "123TestCase"},
		{"make$it#camel%case", "makeItCamelCase"},
		{"test_with_123_numbers", "testWith123Numbers"},
		{"123_456_789", "123456789"},

		// Edge cases
		{"", ""},                              // Empty string
		{"_", ""},                             // Single underscore
		{"---", ""},                           // Only delimiters
		{"___hello___world___", "helloWorld"}, // Leading/trailing underscores
		{"A", "a"},                            // Single character
		{"  spaced  out  ", "spacedOut"},      // Extra spaces
	}

	for _, test := range tests {
		output, _ := ToCamelCase(test.input)
		if output != test.expected {
			t.Errorf("For input '%s', expected '%s' but got '%s'", test.input, test.expected, output)
		}
	}
}
