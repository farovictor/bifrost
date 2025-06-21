package tests

import (
	"time"

	"github.com/farovictor/bifrost/pkg/auth"
)

// makeToken generates a short-lived auth token for a user.
func makeToken(userID string) string {
	t := auth.AuthToken{
		UserID:    userID,
		ExpiresAt: time.Now().Add(time.Hour),
	}
	tok, _ := auth.Sign(t)
	return tok
}
