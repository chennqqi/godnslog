package models

import (
	"crypto/rand"
	"encoding/base32"
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
