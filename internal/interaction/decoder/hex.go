package decoder

import (
	"encoding/hex"
)

// hexDetect checks whether input looks like a hex-encoded string.
// A valid hex string has even length and contains only [0-9a-fA-F].
func hexDetect(input string) bool {
	if len(input) < 2 || len(input)%2 != 0 {
		return false
	}
	for i := 0; i < len(input); i++ {
		c := input[i]
		switch {
		case c >= '0' && c <= '9':
		case c >= 'a' && c <= 'f':
		case c >= 'A' && c <= 'F':
		default:
			return false
		}
	}
	return true
}

// decodeHex decodes a hex string and validates that the result is
// readable text.
func decodeHex(input string) (string, bool) {
	if !hexDetect(input) {
		return "", false
	}
	decoded, err := hex.DecodeString(input)
	if err != nil {
		return "", false
	}
	if !isReadableText(decoded) {
		return "", false
	}
	return string(decoded), true
}

// isReadableText checks whether data consists of readable text.
// It rejects null bytes, 0xFF, and control characters except
// tab, newline, and carriage return.
func isReadableText(data []byte) bool {
	if len(data) == 0 {
		return false
	}
	for _, b := range data {
		switch {
		case b == 0, b == 0xFF:
			return false
		case b < 0x20 && b != '\t' && b != '\n' && b != '\r':
			return false
		case b > 0x7E:
			return false
		}
	}
	return true
}
