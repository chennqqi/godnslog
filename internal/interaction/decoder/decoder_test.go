package decoder

import (
	"testing"
)

// ---------- base64 ----------

func TestDecodeBase64(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantOK  bool
		wantVal string
	}{
		{
			name:    "valid base64 with padding",
			input:   "eyJhZG1pbiI6InRydWUifQ==",
			wantOK:  true,
			wantVal: `{"admin":"true"}`,
		},
		{
			name:    "valid base64 without padding",
			input:   "aGVsbG8",
			wantOK:  true,
			wantVal: "hello",
		},
		{
			name:   "not base64",
			input:  "!!!not-base64!!!",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := decodeBase64(tt.input)
			if ok != tt.wantOK {
				t.Errorf("decodeBase64(%q) ok=%v, want %v", tt.input, ok, tt.wantOK)
				return
			}
			if ok && got != tt.wantVal {
				t.Errorf("decodeBase64(%q) = %q, want %q", tt.input, got, tt.wantVal)
			}
		})
	}
}

// ---------- hex ----------

func TestDecodeHex(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantOK  bool
		wantVal string
	}{
		{
			name:    "valid hex readable",
			input:   "68656c6c6f",
			wantOK:  true,
			wantVal: "hello",
		},
		{
			name:   "valid hex but unreadable",
			input:  "deadbeef",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := decodeHex(tt.input)
			if ok != tt.wantOK {
				t.Errorf("decodeHex(%q) ok=%v, want %v", tt.input, ok, tt.wantOK)
				return
			}
			if ok && got != tt.wantVal {
				t.Errorf("decodeHex(%q) = %q, want %q", tt.input, got, tt.wantVal)
			}
		})
	}
}

// ---------- base32 ----------

func TestDecodeBase32(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantOK  bool
		wantVal string
	}{
		{
			name:    "valid base32",
			input:   "NBSWY3DP",
			wantOK:  true,
			wantVal: "hello",
		},
		{
			name:   "base32-like but invalid",
			input:  "11111111",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := decodeBase32(tt.input)
			if ok != tt.wantOK {
				t.Errorf("decodeBase32(%q) ok=%v, want %v", tt.input, ok, tt.wantOK)
				return
			}
			if ok && got != tt.wantVal {
				t.Errorf("decodeBase32(%q) = %q, want %q", tt.input, got, tt.wantVal)
			}
		})
	}
}

// ---------- Decode integration ----------

func TestDecode(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantEncoding  string
		wantConfident bool
		wantDecoded   string
	}{
		{
			name:          "base64 input",
			input:         "eyJhZG1pbiI6InRydWUifQ==",
			wantEncoding:  "base64",
			wantConfident: true,
			wantDecoded:   `{"admin":"true"}`,
		},
		{
			name:          "hex input",
			input:         "68656c6c6f",
			wantEncoding:  "hex",
			wantConfident: true,
			wantDecoded:   "hello",
		},
		{
			name:          "base32 input",
			input:         "NBSWY3DP",
			wantEncoding:  "base32",
			wantConfident: true,
			wantDecoded:   "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Decode(tt.input)
			if result.Encoding != tt.wantEncoding {
				t.Errorf("Decode(%q).Encoding = %q, want %q", tt.input, result.Encoding, tt.wantEncoding)
			}
			if result.Confident != tt.wantConfident {
				t.Errorf("Decode(%q).Confident = %v, want %v", tt.input, result.Confident, tt.wantConfident)
			}
			if result.Decoded != tt.wantDecoded {
				t.Errorf("Decode(%q).Decoded = %q, want %q", tt.input, result.Decoded, tt.wantDecoded)
			}
		})
	}
}

// ---------- DecodeAll ----------

func TestDecodeAll(t *testing.T) {
	// Hex input should produce a result with hex encoding.
	results := DecodeAll("68656c6c6f")
	found := false
	for _, r := range results {
		if r.Encoding == "hex" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("DecodeAll(%q) should contain a hex result; got %+v", "68656c6c6f", results)
	}
}

// ---------- extractCandidates ----------

func TestExtractCandidates(t *testing.T) {
	// Mixed content should produce at least one candidate.
	input := "prefix eyJhZG1pbiI6InRydWUifQ== suffix"
	candidates := extractCandidates(input)
	if len(candidates) == 0 {
		t.Errorf("extractCandidates(%q) returned empty slice", input)
	}
}
