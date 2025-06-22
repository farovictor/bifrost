package users

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateAPIKey returns a random hex-encoded API key.
func GenerateAPIKey() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}
