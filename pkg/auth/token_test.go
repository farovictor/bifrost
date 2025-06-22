package auth

import (
	"encoding/base64"
	"errors"
	"os"
	"testing"
	"time"
)

// setTestKey assigns a deterministic signing key for tests.
func setTestKey() {
	signingKey = []byte("0123456789abcdef0123456789abcdef")
}

func TestSignVerify(t *testing.T) {
	setTestKey()
	exp := time.Date(2030, 1, 2, 3, 4, 5, 0, time.UTC)
	tok := AuthToken{UserID: "u1", OrgID: "o1", ExpiresAt: exp}
	raw, err := Sign(tok)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	got, err := Verify(raw)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if got.UserID != tok.UserID || got.OrgID != tok.OrgID || !got.ExpiresAt.Equal(tok.ExpiresAt) {
		t.Fatalf("unexpected token: %#v", got)
	}
}

func TestVerifyInvalidSignature(t *testing.T) {
	setTestKey()
	exp := time.Date(2030, 1, 2, 3, 4, 5, 0, time.UTC)
	raw, err := Sign(AuthToken{UserID: "u", ExpiresAt: exp})
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	signingKey = []byte("abcdef0123456789abcdef0123456789")
	_, err = Verify(raw)
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected invalid token, got %v", err)
	}
}

func TestLoadKeyDecodesBase64(t *testing.T) {
	os.Setenv("BIFROST_SIGNING_KEY", base64.StdEncoding.EncodeToString([]byte("mysecret")))
	key := loadKey()
	if string(key) != "mysecret" {
		t.Fatalf("expected decoded key")
	}
	os.Unsetenv("BIFROST_SIGNING_KEY")
}

func TestVerifyExpiredToken(t *testing.T) {
	setTestKey()
	exp := time.Now().Add(-time.Hour)
	raw, err := Sign(AuthToken{UserID: "u", ExpiresAt: exp})
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	_, err = Verify(raw)
	if !errors.Is(err, ErrExpiredToken) {
		t.Fatalf("expected expired token, got %v", err)
	}
}
