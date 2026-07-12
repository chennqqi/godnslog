package decoder

import (
	"net/url"
	"regexp"
)

// Result describes the outcome of decoding a single input string.
type Result struct {
	Original  string `json:"original"`
	Decoded   string `json:"decoded"`
	Encoding  string `json:"encoding"` // "base64", "base32", "hex", or ""
	Confident bool   `json:"confident"`
}

var base64LikeRE = regexp.MustCompile(`[A-Za-z0-9+/=_-]{8,}`)

// Decode attempts to auto-detect and decode a single encoded string.
// It extracts candidate substrings, tries each decoder in order
// (base64 > base32 > hex), and returns the first confident result.
func Decode(input string) Result {
	candidates := extractCandidates(input)
	for _, candidate := range candidates {
		if result := tryDecoders(candidate); result.Confident {
			return result
		}
	}
	return Result{Original: input, Decoded: "", Encoding: "", Confident: false}
}

// DecodeAll attempts to auto-detect and decode all candidate
// substrings within the input, returning every confident result.
func DecodeAll(input string) []Result {
	candidates := extractCandidates(input)
	var results []Result
	seen := make(map[string]bool)
	for _, candidate := range candidates {
		if result := tryDecoders(candidate); result.Confident {
			if !seen[result.Decoded] {
				seen[result.Decoded] = true
				results = append(results, result)
			}
		}
	}
	if results == nil {
		return []Result{}
	}
	return results
}

// tryDecoders attempts to decode input using each supported encoding
// in order: base64, base32, hex.
func tryDecoders(input string) Result {
	if decoded, ok := decodeBase64(input); ok {
		return Result{Original: input, Decoded: decoded, Encoding: "base64", Confident: true}
	}
	if decoded, ok := decodeBase32(input); ok {
		return Result{Original: input, Decoded: decoded, Encoding: "base32", Confident: true}
	}
	if decoded, ok := decodeHex(input); ok {
		return Result{Original: input, Decoded: decoded, Encoding: "hex", Confident: true}
	}
	return Result{Original: input, Decoded: "", Encoding: "", Confident: false}
}

// extractCandidates collects substrings from input that could be
// encoded data. It returns the original string, any URL-decoded
// variant, and all base64-like segments found via regex.
func extractCandidates(input string) []string {
	seen := make(map[string]bool)
	var candidates []string

	addUnique := func(s string) {
		if !seen[s] {
			seen[s] = true
			candidates = append(candidates, s)
		}
	}

	// 1. Always include the original
	addUnique(input)

	// 2. URL-decode the input (if different)
	decoded := input
	if d, err := url.QueryUnescape(input); err == nil {
		decoded = d
		if decoded != input {
			addUnique(decoded)
		}
	}

	// 3. Extract all base64-like segments from the decoded text
	for _, match := range base64LikeRE.FindAllString(decoded, -1) {
		addUnique(match)
	}

	return candidates
}
