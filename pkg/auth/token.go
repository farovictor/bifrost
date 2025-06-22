package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/farovictor/bifrost/pkg/logging"
)

// AuthToken represents an authentication token for API users.
type AuthToken struct {
	UserID    string    `json:"user_id"`
	OrgID     string    `json:"org_id,omitempty"`
	ExpiresAt time.Time `json:"expires_at"`
}

var signingKey []byte

func init() {
	var err error
	signingKey, err = loadKey()
	if err != nil {
		logging.Logger.Fatal().Err(err).Msg("load signing key")
	}
}

// loadKey returns the signing key from the BIFROST_SIGNING_KEY environment
// variable. If the variable is empty or invalid base64, a random key is
// generated instead.
func loadKey() ([]byte, error) {
	if v := os.Getenv("BIFROST_SIGNING_KEY"); v != "" {
		if b, err := base64.StdEncoding.DecodeString(v); err == nil {
			return b, nil
		}
	}
	return generateKey()
}

// generateKey creates a new random signing key.
func generateKey() ([]byte, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	return b, nil
}

// Sign encodes and signs the token using HMAC-SHA256.
func Sign(t AuthToken) (string, error) {
	payload, err := json.Marshal(t)
	if err != nil {
		return "", err
	}
	mac := hmac.New(sha256.New, signingKey)
	mac.Write(payload)
	sig := mac.Sum(nil)
	return base64.StdEncoding.EncodeToString(payload) + "." +
		base64.StdEncoding.EncodeToString(sig), nil
}

// Verify checks the token signature and expiration.
func Verify(raw string) (AuthToken, error) {
	parts := strings.Split(raw, ".")
	if len(parts) != 2 {
		return AuthToken{}, ErrInvalidToken
	}
	payload, err := base64.StdEncoding.DecodeString(parts[0])
	if err != nil {
		return AuthToken{}, ErrInvalidToken
	}
	sig, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return AuthToken{}, ErrInvalidToken
	}
	mac := hmac.New(sha256.New, signingKey)
	mac.Write(payload)
	if !hmac.Equal(sig, mac.Sum(nil)) {
		return AuthToken{}, ErrInvalidToken
	}
	var t AuthToken
	if err := json.Unmarshal(payload, &t); err != nil {
		return AuthToken{}, ErrInvalidToken
	}
	if time.Now().After(t.ExpiresAt) {
		return AuthToken{}, ErrExpiredToken
	}
	return t, nil
}

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
)
