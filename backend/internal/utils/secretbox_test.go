package utils

import (
	"encoding/base64"
	"errors"
	"testing"
)

func TestSecretboxRoundTrip(t *testing.T) {
	secret := []byte("a-sufficiently-long-secret")

	enc, err := SecretboxEncrypt(secret, "client-secret-value")
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if enc == "client-secret-value" {
		t.Fatal("ciphertext equals plaintext")
	}

	dec, err := SecretboxDecrypt(secret, enc)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if dec != "client-secret-value" {
		t.Fatalf("got %q", dec)
	}
}

func TestSecretboxNonceIsRandom(t *testing.T) {
	secret := []byte("a-sufficiently-long-secret")
	a, _ := SecretboxEncrypt(secret, "same")
	b, _ := SecretboxEncrypt(secret, "same")
	if a == b {
		t.Fatal("expected different ciphertexts for the same plaintext")
	}
}

func TestSecretboxTamperFails(t *testing.T) {
	secret := []byte("a-sufficiently-long-secret")
	enc, _ := SecretboxEncrypt(secret, "value")

	raw, _ := base64.StdEncoding.DecodeString(enc)
	raw[len(raw)-1] ^= 0x01
	tampered := base64.StdEncoding.EncodeToString(raw)

	if _, err := SecretboxDecrypt(secret, tampered); !errors.Is(err, ErrSecretboxDecrypt) {
		t.Fatalf("expected ErrSecretboxDecrypt, got %v", err)
	}
}

func TestSecretboxWrongKeyFails(t *testing.T) {
	enc, _ := SecretboxEncrypt([]byte("a-sufficiently-long-secret"), "value")

	if _, err := SecretboxDecrypt([]byte("another-sufficiently-long-secret"), enc); !errors.Is(err, ErrSecretboxDecrypt) {
		t.Fatalf("expected ErrSecretboxDecrypt, got %v", err)
	}
}

func TestSecretboxGarbageInput(t *testing.T) {
	secret := []byte("a-sufficiently-long-secret")
	for _, in := range []string{"", "!!not-base64!!", "YQ=="} {
		if _, err := SecretboxDecrypt(secret, in); !errors.Is(err, ErrSecretboxDecrypt) {
			t.Errorf("input %q: expected ErrSecretboxDecrypt, got %v", in, err)
		}
	}
}

func TestSecretboxEmptySecret(t *testing.T) {
	if _, err := SecretboxEncrypt(nil, "value"); err == nil {
		t.Fatal("expected error for empty secret")
	}
}
