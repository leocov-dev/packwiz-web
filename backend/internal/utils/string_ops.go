package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"math/big"
)

func GenerateRandomString(length int) string {
	if length < 4 {
		length = 4
	}

	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(fmt.Sprintf("Failed to generate random string: %s", err))
	}

	return base64.URLEncoding.EncodeToString(bytes)[:length]
}

func GenerateLinkToken(length int) string {
	if length < 4 {
		length = 4
	}

	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(fmt.Sprintf("Failed to generate random string: %s", err))
	}

	for i := range bytes {
		bytes[i] = charset[bytes[i]%byte(len(charset))]
	}

	return string(bytes)
}

const (
	passwordLetters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	passwordDigits  = "0123456789"
	passwordSymbols = "!@#$%^&*-_=+"
)

// GeneratePassword returns a cryptographically random password of the given
// length (minimum 12) containing at least one letter and one digit. All
// characters satisfy the user password complexity rules.
func GeneratePassword(length int) string {
	if length < 12 {
		length = 12
	}

	charset := passwordLetters + passwordDigits + passwordSymbols
	pick := func(set string) byte {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(set))))
		if err != nil {
			panic(fmt.Sprintf("Failed to generate random password: %s", err))
		}
		return set[n.Int64()]
	}

	out := make([]byte, length)
	for i := range out {
		out[i] = pick(charset)
	}

	// force one letter and one digit at distinct random positions
	letterPos, err := rand.Int(rand.Reader, big.NewInt(int64(length)))
	if err != nil {
		panic(fmt.Sprintf("Failed to generate random password: %s", err))
	}
	digitPos, err := rand.Int(rand.Reader, big.NewInt(int64(length-1)))
	if err != nil {
		panic(fmt.Sprintf("Failed to generate random password: %s", err))
	}
	l := int(letterPos.Int64())
	d := int(digitPos.Int64())
	if d >= l {
		d++
	}
	out[l] = pick(passwordLetters)
	out[d] = pick(passwordDigits)

	return string(out)
}

// HashPassword hashes a plain-text password using bcrypt
func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}
