package auth

import (
	"crypto/rand"
	"errors"
	"os"
	"testing"
)

type failReader struct{}

func (failReader) Read(b []byte) (int, error) {
	return 0, errors.New("fail")
}

func TestGenerateKeyFailure(t *testing.T) {
	orig := rand.Reader
	rand.Reader = failReader{}
	defer func() { rand.Reader = orig }()

	if _, err := generateKey(); err == nil {
		t.Fatalf("expected error")
	}
}

func TestLoadKeyFailure(t *testing.T) {
	orig := rand.Reader
	rand.Reader = failReader{}
	defer func() { rand.Reader = orig }()
	os.Unsetenv("BIFROST_SIGNING_KEY")

	if _, err := loadKey(); err == nil {
		t.Fatalf("expected error")
	}
}
