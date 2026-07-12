package decoder

import (
	"encoding/base32"
	"strings"
)

// base32Detect checks whether input looks like a base32-encoded string.
// A valid base32 string has length that is a multiple of 8 and
// contains only [A-Z2-7=].
func base32Detect(input string) bool {
	if len(input) < 4 || len(input)%8 != 0 {
		return false
	}
	for i := 0; i < len(input); i++ {
		c := input[i]
		switch {
		case c >= 'A' && c <= 'Z':
		case c >= '2' && c <= '7':
		case c == '=':
		default:
			return false
		}
	}
	return true
}

// decodeBase32 decodes a base32 string using standard encoding,
// with automatic padding, and validates the result is readable text.
func decodeBase32(input string) (string, bool) {
	if !base32Detect(input) {
		return "", false
	}

	// Auto-pad to a multiple of 8
	padded := input
	if rem := len(padded) % 8; rem != 0 {
		padded += strings.Repeat("=", 8-rem)
	}

	decoded, err := base32.StdEncoding.DecodeString(padded)
	if err != nil {
		return "", false
	}
	if !isReadableText(decoded) {
		return "", false
	}
	return string(decoded), true
}
