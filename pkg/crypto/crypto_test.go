package crypto

import (
	"bytes"
	"testing"
)

func testKey() []byte {
	return bytes.Repeat([]byte("k"), 32)
}

func TestEncryptDecryptRoundtrip(t *testing.T) {
	key := testKey()
	plaintext := "sk-supersecretapikey-1234"

	ciphertext, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if string(ciphertext) == plaintext {
		t.Fatal("ciphertext should not equal plaintext")
	}

	got, err := Decrypt(ciphertext, key)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if got != plaintext {
		t.Fatalf("expected %q, got %q", plaintext, got)
	}
}

func TestEncryptProducesUniqueNonces(t *testing.T) {
	key := testKey()
	a, err := Encrypt("same-value", key)
	if err != nil {
		t.Fatal(err)
	}
	b, err := Encrypt("same-value", key)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(a, b) {
		t.Fatal("two encryptions of the same value should differ due to random nonce")
	}
}

func TestDecryptWrongKey(t *testing.T) {
	key := testKey()
	ciphertext, _ := Encrypt("secret", key)

	wrongKey := bytes.Repeat([]byte("x"), 32)
	_, err := Decrypt(ciphertext, wrongKey)
	if err == nil {
		t.Fatal("expected error decrypting with wrong key")
	}
}

func TestDecryptTooShort(t *testing.T) {
	key := testKey()
	_, err := Decrypt([]byte("short"), key)
	if err == nil {
		t.Fatal("expected error for too-short ciphertext")
	}
}
