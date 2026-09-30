package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
)

// ErrSecretboxDecrypt is returned when a value cannot be decrypted, e.g. it
// was tampered with or the key changed.
var ErrSecretboxDecrypt = errors.New("secretbox: unable to decrypt value")

func secretboxAEAD(secret []byte) (cipher.AEAD, error) {
	if len(secret) == 0 {
		return nil, errors.New("secretbox: empty secret")
	}
	key := sha256.Sum256(secret)
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, fmt.Errorf("secretbox: create cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("secretbox: create gcm: %w", err)
	}
	return aead, nil
}

// SecretboxEncrypt encrypts plaintext with AES-256-GCM using a key derived
// (SHA-256) from secret. The result is base64(nonce || ciphertext).
func SecretboxEncrypt(secret []byte, plaintext string) (string, error) {
	aead, err := secretboxAEAD(secret)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("secretbox: generate nonce: %w", err)
	}
	sealed := aead.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

// SecretboxDecrypt reverses SecretboxEncrypt. Any failure (bad encoding,
// tampering, wrong key) returns ErrSecretboxDecrypt.
func SecretboxDecrypt(secret []byte, encoded string) (string, error) {
	aead, err := secretboxAEAD(secret)
	if err != nil {
		return "", err
	}
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil || len(raw) < aead.NonceSize()+aead.Overhead() {
		return "", ErrSecretboxDecrypt
	}
	nonce, ciphertext := raw[:aead.NonceSize()], raw[aead.NonceSize():]
	plain, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", ErrSecretboxDecrypt
	}
	return string(plain), nil
}
