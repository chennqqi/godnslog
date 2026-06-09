package models

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// GenerateID generates a unique ID using base32 encoding
func GenerateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return base32.StdEncoding.EncodeToString(bytes)
}

// GenerateToken generates a unique token for payload tracking
func GenerateToken() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return fmt.Sprintf("%x", bytes)
}

// ComputeDeterministicHash computes a deterministic SHA-256 hash of the given data.
// The data is encoded as canonical JSON (sorted keys) to ensure deterministic output.
// Returns the hex-encoded SHA-256 hash.
func ComputeDeterministicHash(data interface{}) (string, error) {
	// Encode data as canonical JSON with sorted keys
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal data: %w", err)
	}

	// Parse and re-marshal with sorted keys to ensure canonical form
	var canonical interface{}
	if err := json.Unmarshal(jsonBytes, &canonical); err != nil {
		return "", fmt.Errorf("failed to unmarshal data: %w", err)
	}

	canonicalBytes, err := json.Marshal(canonical)
	if err != nil {
		return "", fmt.Errorf("failed to marshal canonical data: %w", err)
	}

	// Compute SHA-256 hash
	hash := sha256.Sum256(canonicalBytes)
	return hex.EncodeToString(hash[:]), nil
}
