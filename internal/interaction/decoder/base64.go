package decoder

import (
	"encoding/base64"
	"strings"
)

// base64Detect checks whether input looks like a base64-encoded string.
// It accepts both standard (A-Za-z0-9+/) and URL-safe (A-Za-z0-9-_)
// alphabets, plus optional padding (=).
func base64Detect(input string) bool {
	if len(input) < 4 {
		return false
	}
	for i := 0; i < len(input); i++ {
		c := input[i]
		switch {
		case c >= 'A' && c <= 'Z':
		case c >= 'a' && c <= 'z':
		case c >= '0' && c <= '9':
		case c == '+' || c == '/':
		case c == '-' || c == '_':
		case c == '=':
		default:
			return false
		}
	}
	return true
}

// decodeBase64 attempts to decode a base64 string using standard and
// URL-safe encodings, with automatic padding fix, and validates the
// result is readable text.
func decodeBase64(input string) (string, bool) {
	if !base64Detect(input) {
		return "", false
	}

	// Try standard encoding first
	decoded, err := base64.StdEncoding.DecodeString(input)
	if err == nil && isReadableText(decoded) {
		return string(decoded), true
	}

	// Try URL-safe encoding
	decoded, err = base64.URLEncoding.DecodeString(input)
	if err == nil && isReadableText(decoded) {
		return string(decoded), true
	}

	// Fix padding and retry with standard encoding
	padded := fixBase64Padding(input)
	if padded != input {
		decoded, err = base64.StdEncoding.DecodeString(padded)
		if err == nil && isReadableText(decoded) {
			return string(decoded), true
		}

		// Fix padding and retry with URL-safe encoding
		decoded, err = base64.URLEncoding.DecodeString(padded)
		if err == nil && isReadableText(decoded) {
			return string(decoded), true
		}
	}

	return "", false
}

// fixBase64Padding adds padding characters to make the input length
// a multiple of 4.
func fixBase64Padding(input string) string {
	if rem := len(input) % 4; rem != 0 {
		return input + strings.Repeat("=", 4-rem)
	}
	return input
}
